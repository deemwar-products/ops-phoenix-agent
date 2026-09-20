package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/deemwar-products/sre-agent/internal/detect"
)

// --- Prompt tests ---

func TestBuildPrompt_Empty(t *testing.T) {
	prompt := BuildPrompt([]detect.ErrorGroup{}, "medium")
	if !strings.Contains(prompt, "No errors detected") {
		t.Error("prompt should mention no errors for empty input")
	}
}

func TestBuildPrompt_WithErrors(t *testing.T) {
	groups := []detect.ErrorGroup{{
		Fingerprint: "fp1",
		Message:     "connection refused to db",
		Container:   "api-1",
		Count:       5,
		FirstSeen:   time.Now().Add(-10 * time.Minute),
		LastSeen:    time.Now().Add(-1 * time.Minute),
	}}
	prompt := BuildPrompt(groups, "medium")
	if !strings.Contains(prompt, "connection refused to db") {
		t.Error("prompt should include error message")
	}
	if !strings.Contains(prompt, "api-1") {
		t.Error("prompt should include container")
	}
	if !strings.Contains(prompt, "5") {
		t.Error("prompt should include count")
	}
}

// --- ParseResponse tests ---

func TestParseResponse_ValidJSON(t *testing.T) {
	raw := `{"findings": [{"error": "db down", "root_cause": "OOM", "affected_service": "db", "suggested_fix": "increase memory", "severity": "P1", "confidence": 0.9}], "summary": "DB OOM"}`
	result, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	if result.Findings[0].RootCause != "OOM" {
		t.Errorf("expected root_cause 'OOM', got %q", result.Findings[0].RootCause)
	}
	if result.Summary != "DB OOM" {
		t.Errorf("expected summary 'DB OOM', got %q", result.Summary)
	}
}

func TestParseResponse_EmptyFindings(t *testing.T) {
	raw := `{"findings": [], "summary": "No errors"}`
	result, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}

func TestParseResponse_NonJSON(t *testing.T) {
	raw := "I cannot determine the root cause from this information alone."
	result, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("parse should not error on non-JSON: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings for non-JSON, got %d", len(result.Findings))
	}
}

func TestParseResponse_JSONWithPreamble(t *testing.T) {
	raw := "Here is my analysis:\n" +
		`{"findings": [{"error": "x", "root_cause": "y", "affected_service": "z", "suggested_fix": "w", "severity": "P2", "confidence": 0.7}], "summary": "test"}`
	result, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
}

// --- Diff prompt + extraction tests ---

func TestBuildDiffPrompt_IncludesErrorContext(t *testing.T) {
	finding := Finding{
		Error:        "nil pointer dereference",
		RootCause:    "uninitialized connection pool",
		AffectedSvc:  "api-gateway",
		SuggestedFix: "add nil check before using pool",
		Severity:     "P1",
		Confidence:   0.85,
	}
	grp := detect.ErrorGroup{
		Message:   "nil pointer dereference",
		Container: "api-1",
		Count:     3,
	}
	prompt := BuildDiffPrompt(finding, grp, []string{"main.go", "pool.go"})

	for _, want := range []string{"nil pointer", "uninitialized", "api-gateway", "main.go", "pool.go"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestBuildDiffPrompt_NoFiles(t *testing.T) {
	finding := Finding{Severity: "P2", SuggestedFix: "fix it"}
	grp := detect.ErrorGroup{Message: "boom"}
	prompt := BuildDiffPrompt(finding, grp, nil)
	// No files listed when repoFiles is empty/nil
	if strings.Contains(prompt, "Known repo files:") {
		t.Error("prompt should not mention repo files when none provided")
	}
	if !strings.Contains(prompt, "Diff:") {
		t.Errorf("prompt missing 'Diff:' section")
	}
}

func TestExtractDiff_FindsDiffStart(t *testing.T) {
	raw := "Here's the patch you wanted:\n\n" +
		"--- a/main.go\n+++ b/main.go\n@@ -1,1 +1,1 @@\n-foo\n+bar\n"
	diff := extractDiff(raw)
	if !strings.HasPrefix(diff, "--- a/main.go") {
		t.Errorf("diff should start with file header, got %q", diff)
	}
	if !strings.Contains(diff, "+bar") {
		t.Error("diff should include the +bar line")
	}
}

func TestExtractDiff_NoDiffHeader(t *testing.T) {
	raw := "Sorry, I couldn't generate a patch for this."
	diff := extractDiff(raw)
	// Falls back to trimmed raw
	if diff != strings.TrimSpace(raw) {
		t.Errorf("expected fallback to raw, got %q", diff)
	}
}

func TestExtractDiff_DiffOnly(t *testing.T) {
	raw := "--- a/x.go\n+++ b/x.go\n@@ -1 +1 @@\n-old\n+new\n"
	diff := extractDiff(raw)
	if diff != raw {
		t.Errorf("expected unchanged diff, got %q", diff)
	}
}

// --- HTTP integration tests ---

func TestCallAnthropic_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "rate limited"}`, http.StatusTooManyRequests)
	}))
	defer server.Close()

	origURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = origURL }()

	a := &Analyzer{client: server.Client()}
	_, err := a.callAnthropic(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error from 429 response")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("expected error to mention 429, got: %v", err)
	}
}

func TestCallAnthropic_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	origURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = origURL }()

	a := &Analyzer{client: server.Client()}
	_, err := a.callAnthropic(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error from invalid JSON response")
	}
}

func TestCallAnthropic_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content": [], "model": "test"}`))
	}))
	defer server.Close()

	origURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = origURL }()

	a := &Analyzer{client: server.Client()}
	_, err := a.callAnthropic(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error from empty content array")
	}
}

func TestCallAnthropic_ValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content": [{"text": "analysis done"}], "model": "test"}`))
	}))
	defer server.Close()

	origURL := apiURL
	apiURL = server.URL
	defer func() { apiURL = origURL }()

	a := &Analyzer{client: server.Client()}
	result, err := a.callAnthropic(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "analysis done" {
		t.Errorf("expected 'analysis done', got %q", result)
	}
}
