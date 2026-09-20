package observability

import (
	"testing"
	"time"
)

func TestParseLokiResponse(t *testing.T) {
	data := []byte(`{
		"status": "success",
		"data": {
			"resultType": "streams",
			"result": [
				{
					"stream": {
						"container": "myapp-api",
						"severity": "error"
					},
					"values": [
						["1700000000000000000", "ERROR: connection refused to database"],
						["1700000060000000000", "ERROR: retry attempt 2 failed"]
					]
				},
				{
					"stream": {
						"container": "myapp-worker",
						"severity": "fatal"
					},
					"values": [
						["1700000030000000000", "FATAL: out of memory"]
					]
				}
			]
		}
	}`)

	logs, err := ParseLokiResponse(data)
	if err != nil {
		t.Fatalf("ParseLokiResponse error: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}
	if logs[0].Container != "myapp-api" {
		t.Errorf("expected container 'myapp-api', got %q", logs[0].Container)
	}
	if logs[0].Message != "ERROR: connection refused to database" {
		t.Errorf("unexpected message: %q", logs[0].Message)
	}
	if logs[2].Container != "myapp-worker" {
		t.Errorf("expected container 'myapp-worker', got %q", logs[2].Container)
	}
	if logs[0].Timestamp.UnixNano() != 1700000000000000000 {
		t.Errorf("unexpected timestamp: %d", logs[0].Timestamp.UnixNano())
	}
}

func TestParseLokiResponseError(t *testing.T) {
	data := []byte(`{"status": "error", "error": "bad query"}`)
	_, err := ParseLokiResponse(data)
	if err == nil {
		t.Fatal("expected error for non-success status")
	}
}

func TestParseLokiResponseMalformed(t *testing.T) {
	data := []byte(`not json`)
	_, err := ParseLokiResponse(data)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseLokiResponseEmpty(t *testing.T) {
	data := []byte(`{"status": "success", "data": {"resultType": "streams", "result": []}}`)
	logs, err := ParseLokiResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected 0 logs, got %d", len(logs))
	}
}

func TestParseLokiResponseBadTimestamp(t *testing.T) {
	data := []byte(`{
		"status": "success",
		"data": {
			"resultType": "streams",
			"result": [{
				"stream": {"container": "test"},
				"values": [["not-a-number", "some message"]]
			}]
		}
	}`)
	logs, err := ParseLokiResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs (bad timestamp skipped), got %d", len(logs))
	}
}

func TestBuildLogQL(t *testing.T) {
	tests := []struct {
		name   string
		config FilterConfig
		want   string
	}{
		{
			name: "container pattern only",
			config: FilterConfig{
				ContainerPatterns: []string{"myapp-*", "api-*"},
			},
			want: `{container=~"myapp-.*|api-.*"}`,
		},
		{
			name: "container + label error pattern",
			config: FilterConfig{
				ContainerPatterns: []string{"myapp-*"},
				ErrorPattern:      `level=~"(?i)error|fatal"`,
			},
			want: `{container=~"myapp-.*",level=~"(?i)error|fatal"}`,
		},
		{
			name: "bare error pattern becomes line filter",
			config: FilterConfig{
				ErrorPattern: "error|panic",
			},
			want: `{} |= "(?i)(error|panic)"`,
		},
		{
			name: "container and host patterns",
			config: FilterConfig{
				ContainerPatterns: []string{"web-*"},
				HostPatterns:      []string{"prod-*"},
			},
			want: `{container=~"web-.*",host=~"prod-.*"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLogQL(tt.config)
			if got != tt.want {
				t.Errorf("BuildLogQL:\n got:  %q\n want: %q", got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		in    string
		want  time.Duration
		isErr bool
	}{
		{"30m", 30 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"5s", 5 * time.Second, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if tt.isErr && err == nil {
				t.Errorf("parseDuration(%q) expected error, got %v", tt.in, got)
			}
			if !tt.isErr && err != nil {
				t.Errorf("parseDuration(%q) unexpected error: %v", tt.in, err)
			}
			if !tt.isErr && got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractStackSlug(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://myorg.grafana.net", "myorg"},
		{"myorg", "myorg"},
		{"https://acme.grafana.net/api/health", "acme"},
		{"prod-stack", "prod-stack"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := extractStackSlug(tt.in)
			if got != tt.want {
				t.Errorf("extractStackSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestErrorLog_RawPreserved(t *testing.T) {
	data := []byte(`{
		"status": "success",
		"data": {
			"resultType": "streams",
			"result": [{
				"stream": {"service": "api"},
				"values": [["1700000000000000000", "the exact error text"]]
			}]
		}
	}`)
	logs, err := ParseLokiResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if logs[0].Raw != logs[0].Message {
		t.Error("Raw and Message should match")
	}
	if logs[0].Labels["service"] != "api" {
		t.Errorf("expected label service=api, got %v", logs[0].Labels)
	}
}

func TestGrafanaTimestampParsing(t *testing.T) {
	// Grafana returns ms timestamps
	ms := "1700000000000"
	ts := parseGrafanaTimestamp(ms)
	if !ts.Equal(time.UnixMilli(1700000000000)) {
		t.Errorf("expected %v, got %v", time.UnixMilli(1700000000000), ts)
	}

	// Nanosecond timestamp (19 digits)
	ns := "1700000000000000000"
	ts2 := parseGrafanaTimestamp(ns)
	expected2 := time.Unix(0, 1700000000000000000)
	if !ts2.Equal(expected2) {
		t.Errorf("expected ns timestamp %v, got %v", expected2, ts2)
	}
}
