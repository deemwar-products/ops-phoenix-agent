package detect_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/deemwar-products/sre-agent/internal/detect"
	"github.com/deemwar-products/sre-agent/internal/observability"
)

func makeLog(msg, container string, ts time.Time) observability.ErrorLog {
	return observability.ErrorLog{
		Timestamp: ts,
		Container: container,
		Message:   msg,
		Raw:       msg,
	}
}

func TestDeduplicate_DistinctMessages(t *testing.T) {
	logs := []observability.ErrorLog{
		makeLog("connection refused", "api-1", time.Now()),
		makeLog("nil pointer dereference", "worker-1", time.Now()),
		makeLog("context deadline exceeded", "scheduler-1", time.Now()),
	}
	groups := detect.Deduplicate(logs)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
}

func TestDeduplicate_DuplicatesCollapse(t *testing.T) {
	now := time.Now()
	logs := []observability.ErrorLog{
		makeLog("connection refused", "api-1", now.Add(-5*time.Minute)),
		makeLog("connection refused", "api-2", now.Add(-3*time.Minute)),
		makeLog("connection refused", "api-1", now.Add(-1*time.Minute)),
	}
	groups := detect.Deduplicate(logs)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Count != 3 {
		t.Fatalf("expected count 3, got %d", groups[0].Count)
	}
	if !groups[0].FirstSeen.Equal(now.Add(-5 * time.Minute)) {
		t.Errorf("first seen mismatch: got %v", groups[0].FirstSeen)
	}
	if !groups[0].LastSeen.Equal(now.Add(-1 * time.Minute)) {
		t.Errorf("last seen mismatch: got %v", groups[0].LastSeen)
	}
}

func TestDeduplicate_CaseInsensitive(t *testing.T) {
	logs := []observability.ErrorLog{
		makeLog("CONNECTION REFUSED", "api-1", time.Now()),
		makeLog("connection refused", "api-2", time.Now()),
	}
	groups := detect.Deduplicate(logs)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}

func TestDeduplicate_EmptyInput(t *testing.T) {
	groups := detect.Deduplicate(nil)
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups for nil input, got %d", len(groups))
	}
	groups = detect.Deduplicate([]observability.ErrorLog{})
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups for empty input, got %d", len(groups))
	}
}

func TestDeduplicate_SortedByRecency(t *testing.T) {
	now := time.Now()
	logs := []observability.ErrorLog{
		makeLog("old error", "a", now.Add(-10*time.Minute)),
		makeLog("new error", "b", now.Add(-1*time.Minute)),
		makeLog("mid error", "c", now.Add(-5*time.Minute)),
	}
	groups := detect.Deduplicate(logs)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Message != "new error" {
		t.Errorf("expected first group to be 'new error', got %q", groups[0].Message)
	}
	if groups[1].Message != "mid error" {
		t.Errorf("expected second group to be 'mid error', got %q", groups[1].Message)
	}
	if groups[2].Message != "old error" {
		t.Errorf("expected third group to be 'old error', got %q", groups[2].Message)
	}
}

func TestValidateTimeWindow(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"30m", true}, {"2h", true}, {"1d", true}, {"5m", true},
		{"", false}, {"abc", false}, {"-5m", false},
	}
	for _, tt := range tests {
		err := detect.ValidateTimeWindow(tt.input)
		if tt.ok && err != nil {
			t.Errorf("ValidateTimeWindow(%q) should pass: %v", tt.input, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("ValidateTimeWindow(%q) should fail", tt.input)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"text", true}, {"json", true}, {"markdown", true}, {"md", true},
		{"xml", false}, {"", true}, {"  TEXT  ", true},
	}
	for _, tt := range tests {
		_, err := detect.ParseFormat(tt.input)
		if tt.valid && err != nil {
			t.Errorf("ParseFormat(%q) should pass: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ParseFormat(%q) should fail", tt.input)
		}
	}
}

func TestDefaultTimeWindow(t *testing.T) {
	if got := detect.DefaultTimeWindow(""); got != "1h" {
		t.Errorf("expected '1h', got %q", got)
	}
	if got := detect.DefaultTimeWindow("30m"); got != "30m" {
		t.Errorf("expected '30m', got %q", got)
	}
}

func TestRenderText(t *testing.T) {
	groups := []detect.ErrorGroup{{
		Fingerprint: "fp1",
		Message:     "connection refused to db",
		Container:   "api-1",
		Count:       5,
		FirstSeen:   time.Now().Add(-10 * time.Minute),
		LastSeen:    time.Now().Add(-1 * time.Minute),
	}}
	out := detect.Render(groups, detect.FormatText, "1h")
	if !strings.Contains(out, "Found 1 unique error") {
		t.Errorf("text output missing count: %s", out)
	}
	if !strings.Contains(out, "api-1") {
		t.Errorf("text output missing container: %s", out)
	}
	if !strings.Contains(out, "x5") {
		t.Errorf("text output missing count per group: %s", out)
	}
}

func TestRenderJSON(t *testing.T) {
	groups := []detect.ErrorGroup{{
		Fingerprint: "fp1",
		Message:     "test error",
		Container:   "svc-1",
		Count:       2,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}}
	out := detect.Render(groups, detect.FormatJSON, "1h")
	var parsed []detect.ErrorGroup
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v\n%s", err, out)
	}
	if len(parsed) != 1 || parsed[0].Container != "svc-1" {
		t.Errorf("JSON parsing mismatch")
	}
}

func TestRenderMarkdown(t *testing.T) {
	groups := []detect.ErrorGroup{{
		Fingerprint: "fp1",
		Message:     "test error",
		Container:   "svc-1",
		Count:       1,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}}
	out := detect.Render(groups, detect.FormatMarkdown, "1h")
	if !strings.Contains(out, "# Error Report") {
		t.Errorf("markdown missing heading: %s", out)
	}
	if !strings.Contains(out, "|") {
		t.Errorf("markdown missing table: %s", out)
	}
}

func TestSummary(t *testing.T) {
	if got := detect.Summary(nil); got != "no errors" {
		t.Errorf("expected 'no errors', got %q", got)
	}
	groups := []detect.ErrorGroup{{
		Message:   "db connection failed",
		Container: "api",
		Count:     10,
	}}
	got := detect.Summary(groups)
	if !strings.Contains(got, "1 unique error") {
		t.Errorf("summary missing count: %s", got)
	}
	if !strings.Contains(got, "db connection failed") {
		t.Errorf("summary missing message: %s", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := detect.Truncate("hello", 10); got != "hello" {
		t.Errorf("short string should not truncate, got %q", got)
	}
	long := strings.Repeat("a", 200)
	got := detect.Truncate(long, 50)
	if len(got) != 50 {
		t.Errorf("expected length 50, got %d: %q", len(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncation suffix '...', got %q", got)
	}
}
