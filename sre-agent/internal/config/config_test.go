package config

import (
	"path/filepath"
	"testing"
)

func TestConfigDirPrecedence(t *testing.T) {
	// Save and restore package state
	oldOverride := OverrideDir
	t.Cleanup(func() { OverrideDir = oldOverride })

	t.Run("default under XDG home", func(t *testing.T) {
		OverrideDir = ""
		t.Setenv("SRE_AGENT_CONFIG_DIR", "")
		t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
		if got := ConfigDir(); got != filepath.Join("/tmp/xdg", "sre-agent") {
			t.Errorf("ConfigDir() = %q", got)
		}
	})

	t.Run("env var beats default", func(t *testing.T) {
		OverrideDir = ""
		t.Setenv("SRE_AGENT_CONFIG_DIR", "/tmp/from-env")
		if got := ConfigDir(); got != "/tmp/from-env" {
			t.Errorf("ConfigDir() = %q", got)
		}
	})

	t.Run("flag override beats env var", func(t *testing.T) {
		OverrideDir = "/tmp/from-flag"
		t.Setenv("SRE_AGENT_CONFIG_DIR", "/tmp/from-env")
		if got := ConfigDir(); got != "/tmp/from-flag" {
			t.Errorf("ConfigDir() = %q", got)
		}
	})

	t.Run("ConfigPath uses ConfigDir", func(t *testing.T) {
		OverrideDir = "/tmp/from-flag"
		if got := ConfigPath(); got != "/tmp/from-flag/config.yaml" {
			t.Errorf("ConfigPath() = %q", got)
		}
	})
}
