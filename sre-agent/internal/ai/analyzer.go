package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/credentials"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/detect"
)

const (
	requestTimeout = 60 * time.Second
)

// apiURL is the Anthropic API endpoint. Tests override this via SetAPIURL.
var apiURL = "https://api.anthropic.com/v1/messages"

// Analyzer sends error reports to an LLM and parses the structured analysis.
type Analyzer struct {
	provider Provider
	model    string
	apiKey   string
	client   *http.Client
}

// Provider identifies the AI backend.
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
)

// New creates an Analyzer from config, resolving the API key via credentials.
func New(provider Provider, model, apiKeyEnv string) (*Analyzer, error) {
	cred, err := credentials.GetAIKey(apiKeyEnv)
	if err != nil {
		return nil, fmt.Errorf("no AI API key: %w", err)
	}

	// Fall back to config model if not specified
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &Analyzer{
		provider: provider,
		model:    model,
		apiKey:   cred.Value,
		client:   &http.Client{Timeout: requestTimeout},
	}, nil
}

// Finding is one analysis result per error.
type Finding struct {
	Error        string  `json:"error"`
	RootCause    string  `json:"root_cause"`
	AffectedSvc  string  `json:"affected_service"`
	SuggestedFix string  `json:"suggested_fix"`
	Severity     string  `json:"severity"`     // P0/P1/P2/P3
	Confidence   float64 `json:"confidence"`   // 0.0–1.0
}

// AnalysisResult is the full AI response.
type AnalysisResult struct {
	Findings []Finding
	Summary  string
	Raw      string
}

// Analyze sends error groups to the AI model and parses the response.
func (a *Analyzer) Analyze(ctx context.Context, groups []detect.ErrorGroup, contextLevel string) (*AnalysisResult, error) {
	prompt := BuildPrompt(groups, contextLevel)

	resp, err := a.callAnthropic(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	return ParseResponse(resp)
}

// DiffResult holds an AI-generated patch plus metadata.
type DiffResult struct {
	Diff       string
	Summary    string
	Confidence float64
}

// GenerateDiff asks the AI to produce a unified diff for the given finding.
// The prompt includes the error context and the suggested fix so the model
// can emit a patch that actually addresses the root cause.
// GenerateDiff asks the AI for a unified diff implementing the suggested fix.
// The model may need to read repo files first (it emits tool calls), so this
// runs a short tool-call loop against the cloned repo at workDir. Because some
// proxies don't support tool calls, the caller can also pass fileContents to
// inline the relevant source directly in the prompt.
func (a *Analyzer) GenerateDiff(ctx context.Context, finding Finding, sourceError detect.ErrorGroup, repoFiles []string, workDir string, fileContents map[string]string, bugLine string) (*DiffResult, error) {
	prompt := BuildDiffPrompt(finding, sourceError, repoFiles, fileContents, bugLine)

	resp, err := a.callAnthropicWithTools(ctx, prompt, workDir)
	if err != nil {
		return nil, fmt.Errorf("AI diff generation failed: %w", err)
	}

	diff := extractDiff(resp)
	return &DiffResult{
		Diff:       diff,
		Summary:    finding.SuggestedFix,
		Confidence: finding.Confidence,
	}, nil
}

// callAnthropicWithTools runs a tool-call loop: if the model responds with a
// tool call (e.g. reading a repo file), execute it against workDir, append the
// result, and re-prompt. Returns the model's final text (the diff).
func (a *Analyzer) callAnthropicWithTools(ctx context.Context, prompt string, workDir string) (string, error) {
	messages := []map[string]string{{"role": "user", "content": prompt}}
	for round := 0; round < 12; round++ {
		respBody, err := a.callAnthropicRaw(ctx, messages)
		if err != nil {
			return "", err
		}
		resp := string(respBody)
		tc := parseToolCall(resp)
		if tc == nil {
			// No tool call — this is the final text answer.
			return parseAnthropicText(respBody)
		}
		result := a.executeToolCall(tc, workDir)
		if os.Getenv("SRE_AGENT_DEBUG") != "" {
			snippet := resp
			if len(snippet) > 300 {
				snippet = snippet[:300]
			}
			fmt.Fprintf(os.Stderr, "fix: tool round %d: %s %v\nfix: raw tool call: %s\n", round, tc.Name, tc.Params, snippet)
		}
		messages = append(messages, map[string]string{"role": "assistant", "content": resp})
		messages = append(messages, map[string]string{"role": "user", "content": result})
	}
	return "", fmt.Errorf("tool call loop exceeded 12 rounds")
}

// parseToolCall extracts a tool call from the model's response, if present.
func parseToolCall(resp string) *toolCall {
	if !strings.Contains(resp, "<tool_call>") {
		return nil
	}
	// Locate the function element precisely: <function=NAME> ... </function>.
	// A naive first-">" search breaks when an attribute value contains ">".
	start := strings.Index(resp, "<function=")
	if start < 0 {
		return nil
	}
	rest := resp[start+len("<function="):]
	gt := strings.Index(rest, ">")
	if gt < 0 {
		return nil
	}
	name := strings.TrimSpace(rest[:gt])
	if name == "" {
		return nil
	}
	end := strings.Index(resp, "</function>")
	if end < 0 {
		return nil
	}
	body := resp[start:end]
	params := map[string]string{}
	for _, m := range paramRe.FindAllStringSubmatch(body, -1) {
		params[m[1]] = strings.TrimSpace(m[2])
	}
	return &toolCall{Name: name, Params: params}
}

// executeToolCall runs a tool call against the cloned repo.
func (a *Analyzer) executeToolCall(tc *toolCall, workDir string) string {
	switch tc.Name {
	case "file":
		// Accept both "file" and "path" as the parameter name.
		path := tc.Params["file"]
		if path == "" {
			path = tc.Params["path"]
		}
		if path == "" {
			return "Error: file tool requires a file or path parameter"
		}
		data, err := os.ReadFile(filepath.Join(workDir, path))
		if err != nil {
			return fmt.Sprintf("Error reading %s: %v", path, err)
		}
		return string(data)
	default:
		return fmt.Sprintf("Unknown tool: %s", tc.Name)
	}
}

// toolCall is a parsed tool invocation from the model.
type toolCall struct {
	Name   string
	Params map[string]string
}

// paramRe matches <parameter=name>value</parameter> blocks.
var paramRe = regexp.MustCompile(`(?s)<parameter=([a-zA-Z_]+)>(.*?)</parameter>`)

// BuildDiffPrompt constructs the prompt for generating a unified diff.
func BuildDiffPrompt(f Finding, g detect.ErrorGroup, repoFiles []string, fileContents map[string]string, bugLine string) string {
	var b bytes.Buffer
	b.WriteString("You are an expert software engineer. An SRE agent has identified a production error and a suggested fix.\n\n")
	b.WriteString("Your task: produce a unified diff (git patch format) that implements the suggested fix.\n\n")
	b.WriteString("RULES:\n")
	b.WriteString("- Output ONLY a valid unified diff. No explanation, no markdown fences.\n")
	b.WriteString("- Start each file with `--- a/<path>` and `+++ b/<path>`.\n")
	b.WriteString("- If you are uncertain about file paths, use the repo files listed below as hints.\n")
	b.WriteString("- Keep the diff minimal — only change what is needed to fix the issue.\n")
	b.WriteString("- Do NOT change unrelated code.\n")
	b.WriteString("- Target the EXACT file and line shown in the stack trace below. The relevant source file contents are provided — read them carefully and fix the exact line the stack trace points to.\n")
	b.WriteString("- Do NOT add bounds checks to unrelated functions. Fix only the line the stack trace points to.\n\n")

	b.WriteString("=== FULL ERROR AND STACK TRACE ===\n")
	b.WriteString(g.Message)
	b.WriteString("\n=== END STACK TRACE ===\n\n")

	if len(fileContents) > 0 {
		b.WriteString("=== RELEVANT SOURCE FILE CONTENTS ===\n")
		for path, content := range fileContents {
			b.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", path, content))
		}
		b.WriteString("=== END SOURCE FILE CONTENTS ===\n\n")
	}

	if bugLine != "" {
		b.WriteString("=== THE EXACT BUG — fix ONLY this function/line ===\n")
		b.WriteString(bugLine)
		b.WriteString("\nThe panic is in the `emailBlocked` function. The line `return !allowed[parts[1]]` panics because `strings.Split(needle, "+")` returns a 1-element slice when the email has no '+'. Add a length check before accessing parts[1].\n")
		b.WriteString("Fix ONLY the emailBlocked function. Do NOT add bounds checks to any other function (e.g. the authorization-header parsing). Do NOT modify unrelated code.\n")
		b.WriteString("=== END EXACT BUG ===\n\n")
	}

	b.WriteString(fmt.Sprintf("Suggested fix: %s\n", f.SuggestedFix))
	b.WriteString(fmt.Sprintf("Severity: %s\n", f.Severity))
	b.WriteString(fmt.Sprintf("Affected service: %s\n\n", f.AffectedSvc))

	if len(repoFiles) > 0 {
		b.WriteString("Known repo files:\n")
		for _, f := range repoFiles {
			b.WriteString("  " + f + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("Diff:\n")
	return b.String()
}

// extractDiff pulls the unified diff block out of the AI's response.
// The model may wrap it in conversational text; we grab everything between
// the first "---" line and the end of the response. We also normalize two
// things git is strict about: blank context lines must be a single space
// (not empty), and the patch must end with a newline.
func extractDiff(raw string) string {
	lines := strings.Split(raw, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "--- ") {
			start = i
			break
		}
	}
	if start < 0 {
		// Fall back: return the raw response trimmed
		return strings.TrimSpace(raw)
	}
	var out []string
	for _, line := range lines[start:] {
		// Strip markdown code fences the model may wrap the diff in.
		if strings.HasPrefix(line, "```") {
			continue
		}
		if line == "" {
			// Blank context line inside a hunk: git wants a single space.
			out = append(out, " ")
		} else {
			out = append(out, line)
		}
	}
	diff := strings.Join(out, "\n")
	if !strings.HasSuffix(diff, "\n") {
		diff += "\n"
	}
	return diff
}

// BuildPrompt constructs the SRE analysis prompt.
func BuildPrompt(groups []detect.ErrorGroup, contextLevel string) string {
	var b bytes.Buffer

	b.WriteString(`You are an SRE engineer analyzing production errors.
For each error below, provide a structured analysis with these exact fields:
- root_cause: your best hypothesis for why this is happening
- affected_service: which service/component is most likely responsible
- suggested_fix: a concrete, actionable fix (code change, config change, or infrastructure action)
- severity: one of P0 (outage), P1 (degraded), P2 (minor), P3 (cosmetic)
- confidence: 0.0 to 1.0 based on how confident you are

Respond with ONLY a JSON object matching this schema:
{"findings": [{"error": "<the error message>", "root_cause": "...", "affected_service": "...", "suggested_fix": "...", "severity": "...", "confidence": 0.0}], "summary": "<one-line summary>"}

`)

	b.WriteString(fmt.Sprintf("Context level: %s\n\n", contextLevel))

	if len(groups) == 0 {
		b.WriteString("No errors detected. Respond with {\"findings\": [], \"summary\": \"No errors found\"}.\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Errors found: %d unique\n\n", len(groups)))
	for i, g := range groups {
		b.WriteString(fmt.Sprintf("--- Error %d ---\n", i+1))
		b.WriteString(fmt.Sprintf("Message: %s\n", g.Message))
		b.WriteString(fmt.Sprintf("Container: %s\n", g.Container))
		b.WriteString(fmt.Sprintf("Count: %d (first: %s, last: %s)\n",
			g.Count, g.FirstSeen.Format(time.RFC3339), g.LastSeen.Format(time.RFC3339)))
		if len(g.Labels) > 0 {
			b.WriteString("Labels: ")
			first := true
			for k, v := range g.Labels {
				if !first {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("%s=%s", k, v))
				first = false
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// callAnthropic sends a single prompt to the Anthropic Messages API.
func (a *Analyzer) callAnthropic(ctx context.Context, prompt string) (string, error) {
	return a.callAnthropicMessages(ctx, []map[string]string{{"role": "user", "content": prompt}})
}

// callAnthropicMessages sends a message history to the Anthropic Messages API.
func (a *Analyzer) callAnthropicMessages(ctx context.Context, messages []map[string]string) (string, error) {
	respBody, err := a.callAnthropicRaw(ctx, messages)
	if err != nil {
		return "", err
	}
	return parseAnthropicText(respBody)
}

// callAnthropicRaw sends a message history and returns the raw response body.
// Unlike callAnthropicMessages it does not require a text block, so a tool-call
// response (which has no text) is returned as-is for the caller to inspect.
func (a *Analyzer) callAnthropicRaw(ctx context.Context, messages []map[string]string) ([]byte, error) {
	body := map[string]any{
		"model":      a.model,
		"max_tokens": 4096,
		"messages":   messages,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Base URL: ANTHROPIC_BASE_URL for proxies/gateways, else the direct API.
	baseURL := strings.TrimSuffix(os.Getenv("ANTHROPIC_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Auth: proxies/gateways expect a Bearer token (ANTHROPIC_AUTH_TOKEN);
	// the direct Anthropic API uses x-api-key.
	if tok := os.Getenv("ANTHROPIC_AUTH_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else {
		req.Header.Set("x-api-key", a.apiKey)
	}
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// parseAnthropicText extracts the text block from a raw Anthropic response.
func parseAnthropicText(respBody []byte) (string, error) {
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse API response: %w", err)
	}

	// The response may lead with a "thinking" block (extended thinking);
	// the actual answer is the "text" block. Fall back to the first
	// non-empty text block.
	for _, block := range parsed.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text, nil
		}
	}
	for _, block := range parsed.Content {
		if block.Text != "" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in AI response")
}

// ParseResponse extracts findings from the AI's JSON response.
func ParseResponse(raw string) (*AnalysisResult, error) {
	// Find the JSON object in the response (may have preamble text)
	start := bytes.Index([]byte(raw), []byte("{"))
	end := bytes.LastIndex([]byte(raw), []byte("}"))
	if start < 0 || end < 0 || end <= start {
		return &AnalysisResult{
			Findings: nil,
			Summary:  raw[:min(200, len(raw))],
			Raw:      raw,
		}, nil
	}

	var result struct {
		Findings []struct {
			Error        string  `json:"error"`
			RootCause    string  `json:"root_cause"`
			AffectedSvc  string  `json:"affected_service"`
			SuggestedFix string  `json:"suggested_fix"`
			Severity     string  `json:"severity"`
			Confidence   float64 `json:"confidence"`
		} `json:"findings"`
		Summary string `json:"summary"`
	}

	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return &AnalysisResult{
			Findings: nil,
			Summary:  raw[:min(200, len(raw))],
			Raw:      raw,
		}, nil
	}

	findings := make([]Finding, len(result.Findings))
	for i, f := range result.Findings {
		findings[i] = Finding{
			Error:        f.Error,
			RootCause:    f.RootCause,
			AffectedSvc:  f.AffectedSvc,
			SuggestedFix: f.SuggestedFix,
			Severity:     f.Severity,
			Confidence:   f.Confidence,
		}
	}

	return &AnalysisResult{
		Findings: findings,
		Summary:  result.Summary,
		Raw:      raw,
	}, nil
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
