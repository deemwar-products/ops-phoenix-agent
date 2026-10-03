package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/detect"
)

// TeamsNotifier posts Adaptive Cards to a Teams channel via a Power Automate
// workflow webhook ("When a Teams webhook request is received" trigger).
type TeamsNotifier struct {
	webhookURL string
	maxShown   int
	client     *http.Client
}

// NewTeamsNotifier builds a notifier; maxShown caps error groups per card.
func NewTeamsNotifier(webhookURL string, maxShown int) *TeamsNotifier {
	if maxShown <= 0 {
		maxShown = 5
	}
	return &TeamsNotifier{
		webhookURL: webhookURL,
		maxShown:   maxShown,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *TeamsNotifier) Name() string { return "teams" }

// Notify builds the card and POSTs it. Power Automate answers 202 on success.
func (t *TeamsNotifier) Notify(ctx context.Context, r Result) error {
	payload := map[string]any{
		"type": "message",
		"attachments": []map[string]any{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"contentUrl":  nil,
				"content":     buildCard(r, t.maxShown),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal adaptive card: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("teams webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("teams webhook → HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}
	return nil
}

// buildCard renders the Result as an Adaptive Card (v1.4).
func buildCard(r Result, maxShown int) map[string]any {
	var icon, title, style string
	switch r.Status {
	case StatusSuccess:
		icon, title, style = "✅", "SRE Agent — all clear", "good"
	case StatusErrors:
		icon, title, style = "🔴", fmt.Sprintf("SRE Agent — %d unique error(s)", len(r.Groups)), "attention"
	default:
		icon, title, style = "⚠️", "SRE Agent — run failed", "warning"
	}

	bodyItems := []map[string]any{
		{
			"type":  "Container",
			"style": style,
			"items": []map[string]any{
				{"type": "TextBlock", "text": fmt.Sprintf("%s %s", icon, title), "weight": "Bolder", "size": "Medium"},
			},
		},
	}

	facts := []map[string]any{}
	if r.Environment != "" {
		facts = append(facts, map[string]any{"title": "Environment", "value": r.Environment})
	}
	facts = append(facts,
		map[string]any{"title": "Command", "value": r.Command},
		map[string]any{"title": "Window", "value": r.TimeWindow},
		map[string]any{"title": "Duration", "value": r.Duration.Round(time.Second).String()},
	)
	bodyItems = append(bodyItems, map[string]any{"type": "FactSet", "facts": facts})

	switch r.Status {
	case StatusErrors:
		shown := r.Groups
		if len(shown) > maxShown {
			shown = shown[:maxShown]
		}
		for _, g := range shown {
			bodyItems = append(bodyItems, map[string]any{
				"type": "TextBlock",
				"text": fmt.Sprintf("**[%s]** (x%d) %s", g.Container, g.Count, detect.Truncate(g.Message, 140)),
				"wrap": true,
			})
		}
		if extra := len(r.Groups) - len(shown); extra > 0 {
			bodyItems = append(bodyItems, map[string]any{
				"type": "TextBlock", "text": fmt.Sprintf("…and %d more", extra), "isSubtle": true, "wrap": true,
			})
		}
	case StatusFailed:
		bodyItems = append(bodyItems, map[string]any{"type": "TextBlock", "text": r.Err, "wrap": true})
	default:
		bodyItems = append(bodyItems, map[string]any{
			"type": "TextBlock", "text": "No errors found in this window.", "isSubtle": true, "wrap": true,
		})
	}

	card := map[string]any{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.4",
		"body":    bodyItems,
	}
	if r.LogsURL != "" {
		card["actions"] = []map[string]any{
			{"type": "Action.OpenUrl", "title": "View logs in Grafana", "url": r.LogsURL},
		}
	}
	return card
}
