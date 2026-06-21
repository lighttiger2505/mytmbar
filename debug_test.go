package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugLogPath(t *testing.T) {
	home := t.TempDir()

	tests := []struct {
		name       string
		configFile string
		xdgState   string
		wantSuffix string
	}{
		{
			name:       "config File override is returned as-is",
			configFile: "/custom/path/debug.log",
			wantSuffix: "/custom/path/debug.log",
		},
		{
			name:       "XDG_STATE_HOME is used when File is empty",
			xdgState:   filepath.Join(home, "state"),
			wantSuffix: filepath.Join(home, "state", "mytmbar", "debug.log"),
		},
		{
			name:       "fallback to ~/.local/state when XDG_STATE_HOME is unset",
			wantSuffix: filepath.Join(".local", "state", "mytmbar", "debug.log"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", tt.xdgState)
			cfg := defaultConfig()
			cfg.Debug.File = tt.configFile

			got := debugLogPath(cfg)

			if tt.configFile != "" {
				if got != tt.configFile {
					t.Errorf("debugLogPath() = %q, want %q", got, tt.configFile)
				}
				return
			}
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("debugLogPath() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestSaveClaudeDebug(t *testing.T) {
	idleContent := "✳ Claude\n❯ "
	runningContent := "· Doing something... (3s · esc to interrupt)"

	tests := []struct {
		name        string
		paneID      string
		content     string
		wantState   string
		wantMode    string
		wantContent string
	}{
		{
			name:        "idle state is recorded",
			paneID:      "%0",
			content:     idleContent,
			wantState:   "Idle",
			wantMode:    "",
			wantContent: idleContent,
		},
		{
			name:        "running state is recorded",
			paneID:      "%1",
			content:     runningContent,
			wantState:   "Running",
			wantMode:    "",
			wantContent: runningContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "debug.log")
			cfg := defaultConfig()
			cfg.Debug.Enabled = true
			cfg.Debug.File = logPath

			saveClaudeDebug(tt.paneID, tt.content, cfg)

			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("log file was not created: %v", err)
			}
			body := string(data)

			if !strings.Contains(body, "pane="+tt.paneID) {
				t.Errorf("log does not contain pane=%s:\n%s", tt.paneID, body)
			}
			if !strings.Contains(body, "state="+tt.wantState) {
				t.Errorf("log does not contain state=%s:\n%s", tt.wantState, body)
			}
			if tt.wantMode != "" && !strings.Contains(body, "mode="+tt.wantMode) {
				t.Errorf("log does not contain mode=%s:\n%s", tt.wantMode, body)
			}
			if !strings.Contains(body, tt.wantContent) {
				t.Errorf("log does not contain captured content:\n%s", body)
			}
		})
	}
}

func TestSaveClaudeDebug_Append(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "debug.log")
	cfg := defaultConfig()
	cfg.Debug.Enabled = true
	cfg.Debug.File = logPath

	saveClaudeDebug("%0", "first call content", cfg)
	saveClaudeDebug("%0", "second call content", cfg)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not found: %v", err)
	}
	body := string(data)

	count := strings.Count(body, "=====")
	// Each block has an opening "=====" line, so expect at least 2.
	if count < 2 {
		t.Errorf("expected at least 2 header blocks, got %d:\n%s", count, body)
	}
	if !strings.Contains(body, "first call content") {
		t.Errorf("first call content missing from log:\n%s", body)
	}
	if !strings.Contains(body, "second call content") {
		t.Errorf("second call content missing from log:\n%s", body)
	}
}
