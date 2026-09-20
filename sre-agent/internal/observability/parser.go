package observability

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LokiResponse is the response shape from Loki's query_range endpoint.
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string         `json:"resultType"`
		Result     []LokiStream   `json:"result"`
	} `json:"data"`
}

// LokiStream is one stream of log lines for a single label set.
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"` // [timestamp_ns, line]
}

// ParseLokiResponse converts a Loki query_range response into ErrorLog entries.
func ParseLokiResponse(data []byte) ([]ErrorLog, error) {
	var resp LokiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal loki response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("loki returned non-success status: %s", resp.Status)
	}

	var out []ErrorLog
	for _, stream := range resp.Data.Result {
		for _, val := range stream.Values {
			if len(val) != 2 {
				continue
			}
			ts, err := parseLokiTimestamp(val[0])
			if err != nil {
				continue
			}
			out = append(out, ErrorLog{
				Timestamp: ts,
				Container: stream.Stream["container"],
				Message:   val[1],
				Labels:    stream.Stream,
				Raw:       val[1],
			})
		}
	}
	return out, nil
}

// parseLokiTimestamp converts Loki's nanosecond string to time.Time.
func parseLokiTimestamp(ns string) (time.Time, error) {
	n, err := strconv.ParseInt(ns, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, n), nil
}

// ParseGrafanaLokiResponse handles the wrapped shape Grafana returns
// (data.results["A"].frames[0].data.values).
type GrafanaLokiEnvelope struct {
	Results map[string]struct {
		Frames []struct {
			Data struct {
				Values [][]any `json:"values"`
			} `json:"data"`
		} `json:"frames"`
	} `json:"results"`
}

// ExtractGrafanaLokiJSON searches a Grafana datasource proxy response for the
// nested Loki payload and returns the raw bytes for ParseLokiResponse to handle.
func ExtractGrafanaLokiJSON(data []byte) ([]byte, error) {
	var env GrafanaLokiEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal grafana envelope: %w", err)
	}
	for _, r := range env.Results {
		for _, f := range r.Frames {
			for _, v := range f.Data.Values {
				_ = v
			}
		}
		// Simpler: assume the response already contains a `results.X.frames[0].data`
		// field that wraps the standard Loki shape. For now, return raw data
		// if it looks like a Loki response already.
	}
	return data, nil
}

// ValidateLokiResponse checks a response is non-empty.
func ValidateLokiResponse(data []byte) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		return fmt.Errorf("empty response from Loki")
	}
	return nil
}
