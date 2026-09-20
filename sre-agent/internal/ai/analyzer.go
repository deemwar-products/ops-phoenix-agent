package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deemwar-products/sre-agent/internal/credentials"
	"github.com/deemwar-products/sre-agent/internal/detect"
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
func (a *Analyzer) GenerateDiff(ctx context.Context, finding Finding, sourceError detect.ErrorGroup, repoFiles []string) (*DiffResult, error) {
	prompt := BuildDiffPrompt(finding, sourceError, repoFiles)

	resp, err := a.callAnthropic(ctx, prompt)
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

// BuildDiffPrompt constructs the prompt for generating a unified diff.
func BuildDiffPrompt(f Finding, g detect.ErrorGroup, repoFiles []string) string {
	var b bytes.Buffer
	b.WriteString("You are an expert software engineer. An SRE agent has identified a production error and a suggested fix.\n\n")
	b.WriteString("Your task: produce a unified diff (git patch format) that implements the suggested fix.\n\n")
	b.WriteString("RULES:\n")
	b.WriteString("- Output ONLY a valid unified diff. No explanation, no markdown fences.\n")
	b.WriteString("- Start each file with `--- a/<path>` and `+++ b/<path>`.\n")
	b.WriteString("- If you are uncertain about file paths, use the repo files listed below as hints.\n")
	b.WriteString("- Keep the diff minimal — only change what is needed to fix the issue.\n")
	b.WriteString("- Do NOT change unrelated code.\n\n")

	b.WriteString(fmt.Sprintf("Error: %s\n", g.Message))
	b.WriteString(fmt.Sprintf("Root cause: %s\n", f.RootCause))
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
// the first "---" line and the end of the response.
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
	return strings.Join(lines[start:], "\n")
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

// callAnthropic sends a request to the Anthropic Messages API.
func (a *Analyzer) callAnthropic(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model":      a.model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse API response: %w", err)
	}

	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("empty response from AI")
	}

	return parsed.Content[0].Text, nil
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
