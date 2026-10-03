package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/detect"
)

// captureServer records the last POST body and replies with the given status.
func captureServer(t *testing.T, status int, captured *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(captured); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(status)
	}))
}

// cardFromPayload unwraps attachments[0].content from the Teams envelope.
func cardFromPayload(t *testing.T, p map[string]any) map[string]any {
	t.Helper()
	if p["type"] != "message" {
		t.Errorf(`envelope type = %v, want "message"`, p["type"])
	}
	atts, ok := p["attachments"].([]any)
	if !ok || len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %v", p["attachments"])
	}
	att := atts[0].(map[string]any)
	if att["contentType"] != "application/vnd.microsoft.card.adaptive" {
		t.Errorf("contentType = %v", att["contentType"])
	}
	return att["content"].(map[string]any)
}

// bodyText flattens every TextBlock/Container text in the card for assertions.
func bodyText(card map[string]any) string {
	var sb strings.Builder
	var walk func(items []any)
	walk = func(items []any) {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if txt, ok := m["text"].(string); ok {
				sb.WriteString(txt + "\n")
			}
			if subs, ok := m["items"].([]any); ok {
				walk(subs)
			}
		}
	}
	if body, ok := card["body"].([]any); ok {
		walk(body)
	}
	return sb.String()
}

func TestTeamsNotifySuccessCard(t *testing.T) {
	var payload map[string]any
	srv := captureServer(t, http.StatusAccepted, &payload)
	defer srv.Close()

	n := NewTeamsNotifier(srv.URL, 5)
	err := n.Notify(context.Background(), Result{
		Command:     "detect",
		Status:      StatusSuccess,
		Environment: "reqsume-prod",
		TimeWindow:  "1h",
		Duration:    2 * time.Second,
		LogsURL:     "https://observability.deemwar.com/d/reqsume-logs",
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	card := cardFromPayload(t, payload)
	if card["version"] != "1.4" {
		t.Errorf("card version = %v", card["version"])
	}
	text := bodyText(card)
	for _, want := range []string{"all clear", "No errors found"} {
		if !strings.Contains(text, want) {
			t.Errorf("card missing %q\ncard: %s", want, text)
		}
	}
	if _, ok := card["actions"]; !ok {
		t.Error("expected OpenUrl action when LogsURL is set")
	}
}

func TestTeamsNotifyErrorsCardTruncates(t *testing.T) {
	var payload map[string]any
	srv := captureServer(t, http.StatusOK, &payload)
	defer srv.Close()

	groups := make([]detect.ErrorGroup, 7)
	for i := range groups {
		groups[i] = detect.ErrorGroup{Message: "boom", Container: "api", Count: i + 1}
	}

	n := NewTeamsNotifier(srv.URL, 5)
	err := n.Notify(context.Background(), Result{
		Command: "detect", Status: StatusErrors, TimeWindow: "1h", Groups: groups,
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	text := bodyText(cardFromPayload(t, payload))
	if !strings.Contains(text, "7 unique error(s)") {
		t.Errorf("title should report total count\ncard: %s", text)
	}
	if !strings.Contains(text, "…and 2 more") {
		t.Errorf("expected truncation note for 7 groups with max 5\ncard: %s", text)
	}
	if strings.Count(text, "**[api]**") != 5 {
		t.Errorf("expected exactly 5 error blocks\ncard: %s", text)
	}
}

func TestTeamsNotifyFailedCard(t *testing.T) {
	var payload map[string]any
	srv := captureServer(t, http.StatusAccepted, &payload)
	defer srv.Close()

	n := NewTeamsNotifier(srv.URL, 5)
	err := n.Notify(context.Background(), Result{
		Command: "detect", Status: StatusFailed, TimeWindow: "1h",
		Err: "backend unreachable: HTTP 521",
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	text := bodyText(cardFromPayload(t, payload))
	for _, want := range []string{"run failed", "HTTP 521"} {
		if !strings.Contains(text, want) {
			t.Errorf("card missing %q\ncard: %s", want, text)
		}
	}
}

func TestTeamsNotifyHTTPError(t *testing.T) {
	var payload map[string]any
	srv := captureServer(t, http.StatusInternalServerError, &payload)
	defer srv.Close()

	n := NewTeamsNotifier(srv.URL, 5)
	err := n.Notify(context.Background(), Result{Command: "detect", Status: StatusSuccess, TimeWindow: "1h"})
	if err == nil {
		t.Fatal("expected error on HTTP 500")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("error should mention status code: %v", err)
	}
}
