package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deemwar-products/sre-agent/internal/history"
)

// withTempHistoryDir sets XDG_CONFIG_HOME to a temp directory so history
// reads/writes don't pollute the real config. Restores original on cleanup.
func withTempHistoryDir(t *testing.T) string {
	t.Helper()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	dir := t.TempDir()
	if err := os.Setenv("XDG_CONFIG_HOME", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if origXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})
	return dir
}

func TestAppendAndRecent(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	_ = history.Append(history.Run{Command: "detect", Status: "success", StartedAt: time.Now(), EndedAt: time.Now()})
	_ = history.Append(history.Run{Command: "analyze", Status: "partial", StartedAt: time.Now(), EndedAt: time.Now()})

	runs, err := history.Recent(10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	if runs[0].Command != "detect" {
		t.Errorf("first run command = %q", runs[0].Command)
	}
	if runs[1].Command != "analyze" {
		t.Errorf("second run command = %q", runs[1].Command)
	}
}

func TestAppend_CapsAtMax(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	now := time.Now()
	for i := 0; i < history.MaxRecentRuns+5; i++ {
		_ = history.Append(history.Run{Command: "detect", Status: "success", StartedAt: now, EndedAt: now})
	}
	runs, _ := history.Recent(100)
	if len(runs) != history.MaxRecentRuns {
		t.Errorf("expected capped at %d, got %d", history.MaxRecentRuns, len(runs))
	}
}

func TestRecent_NilLog(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()
	runs, err := history.Recent(5)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for fresh state, got %d", len(runs))
	}
}

func TestSaveAndLoad(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	now := time.Now().UTC()
	log := &history.RunLog{
		Runs: []history.Run{
			{Command: "fix", Status: "success", Summary: "patched handler.go", StartedAt: now, EndedAt: now},
		},
	}
	if err := history.Save(log); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := history.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(loaded.Runs))
	}
	if loaded.Runs[0].Command != "fix" {
		t.Errorf("command = %q", loaded.Runs[0].Command)
	}
	if loaded.Runs[0].Summary != "patched handler.go" {
		t.Errorf("summary = %q", loaded.Runs[0].Summary)
	}
}

func TestReset(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	_ = history.Append(history.Run{Command: "detect", Status: "success"})
	runs, _ := history.Recent(10)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	_ = history.Reset()
	runs, _ = history.Recent(10)
	if len(runs) != 0 {
		t.Errorf("expected 0 runs after reset, got %d", len(runs))
	}
}

func TestRun_Duration(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	started := time.Now().UTC()
	r := history.Run{Command: "detect", Status: "success", StartedAt: started, EndedAt: started.Add(3 * time.Second), Duration: "3s"}
	if r.Duration != "3s" {
		t.Errorf("Duration = %q, want %q", r.Duration, "3s")
	}
}

func TestRecent_Limit(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	now := time.Now()
	for i := 0; i < 5; i++ {
		_ = history.Append(history.Run{Command: "detect", Status: "success", StartedAt: now, EndedAt: now})
	}
	runs, _ := history.Recent(3)
	if len(runs) != 3 {
		t.Errorf("expected 3, got %d", len(runs))
	}
}

func TestPath_ReturnsConsistentPath(t *testing.T) {
	path := history.Path()
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	if filepath.Base(path) != "runs.yaml" {
		t.Errorf("expected runs.yaml, got %q", filepath.Base(path))
	}
}

func TestAppend_ErrorStatus(t *testing.T) {
	withTempHistoryDir(t)
	_ = history.Reset()

	_ = history.Append(history.Run{
		Command: "fix",
		Status:  "failed",
		Error:   "connection refused",
		StartedAt: time.Now(), EndedAt: time.Now(),
	})
	runs, _ := history.Recent(1)
	if runs[0].Error != "connection refused" {
		t.Errorf("Error = %q", runs[0].Error)
	}
}
