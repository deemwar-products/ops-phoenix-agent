package observability

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is the HTTP client used by all adapters.
var HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// doGET performs a GET with optional headers and returns the body.
func doGET(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return do(req)
}

// doPOST performs a POST with optional body and returns the response.
func doPOST(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if _, ok := headers["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return do(req)
}

func do(req *http.Request) ([]byte, error) {
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s → HTTP %d: %s",
			req.Method, req.URL.Path, resp.StatusCode, truncate(string(body), 200))
	}

	return body, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
