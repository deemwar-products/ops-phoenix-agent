package cli

import "testing"

func TestScanConfigDirArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no flag", []string{"detect"}, ""},
		{"short flag with space", []string{"-c", "/tmp/x", "detect"}, "/tmp/x"},
		{"long flag with space", []string{"--config", "/tmp/x", "status"}, "/tmp/x"},
		{"short flag with equals", []string{"-c=/tmp/x", "detect"}, "/tmp/x"},
		{"long flag with equals", []string{"--config=/tmp/x", "detect"}, "/tmp/x"},
		{"flag after subcommand", []string{"detect", "-c", "/tmp/x"}, "/tmp/x"},
		{"flag without value", []string{"detect", "-c"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanConfigDirArg(tt.args); got != tt.want {
				t.Errorf("scanConfigDirArg(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
