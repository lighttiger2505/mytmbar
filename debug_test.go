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

func TestSaveClaudeDebug_StateFilter(t *testing.T) {
	idleContent := "✳ Claude\n❯ "
	runningContent := "· Doing something... (3s · esc to interrupt)"

	tests := []struct {
		name         string
		content      string
		filterStates []string
		wantWritten  bool
	}{
		{
			name:         "idle passes when idle is in filter",
			content:      idleContent,
			filterStates: []string{"Idle", "Unknown"},
			wantWritten:  true,
		},
		{
			name:         "running is dropped when not in filter",
			content:      runningContent,
			filterStates: []string{"Idle", "Unknown"},
			wantWritten:  false,
		},
		{
			name:         "case-insensitive match works",
			content:      idleContent,
			filterStates: []string{"idle", "unknown"},
			wantWritten:  true,
		},
		{
			name:         "empty filter logs all states",
			content:      runningContent,
			filterStates: nil,
			wantWritten:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "debug.log")
			cfg := defaultConfig()
			cfg.Debug.Enabled = true
			cfg.Debug.File = logPath
			cfg.Debug.States = tt.filterStates

			saveClaudeDebug("%0", tt.content, cfg)

			_, err := os.Stat(logPath)
			exists := err == nil
			if exists != tt.wantWritten {
				t.Errorf("log file exists=%v, want %v", exists, tt.wantWritten)
			}
		})
	}
}

func TestSaveClaudeDebug_SkipConsecutiveDuplicate(t *testing.T) {
	content1 := "✳ Claude\n❯ "
	content2 := "· Doing something... (3s · esc to interrupt)"

	t.Run("same content twice is written only once", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "debug.log")
		cfg := defaultConfig()
		cfg.Debug.Enabled = true
		cfg.Debug.File = logPath

		saveClaudeDebug("%0", content1, cfg)
		saveClaudeDebug("%0", content1, cfg) // duplicate — should be skipped

		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("log file not found: %v", err)
		}
		// Each header line is "===== ... =====" so each block contributes 2 occurrences of "=====".
		count := strings.Count(string(data), "=====")
		if count != 2 {
			t.Errorf("expected 2 (1 block × 2 markers), got %d:\n%s", count, string(data))
		}
	})

	t.Run("different content after duplicate is written", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "debug.log")
		cfg := defaultConfig()
		cfg.Debug.Enabled = true
		cfg.Debug.File = logPath

		saveClaudeDebug("%0", content1, cfg)
		saveClaudeDebug("%0", content1, cfg) // skipped
		saveClaudeDebug("%0", content2, cfg) // different — must be written

		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("log file not found: %v", err)
		}
		count := strings.Count(string(data), "=====")
		if count != 4 {
			t.Errorf("expected 4 (2 blocks × 2 markers), got %d:\n%s", count, string(data))
		}
	})

	t.Run("non-consecutive same content is written again", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "debug.log")
		cfg := defaultConfig()
		cfg.Debug.Enabled = true
		cfg.Debug.File = logPath

		saveClaudeDebug("%0", content1, cfg)
		saveClaudeDebug("%0", content2, cfg) // different
		saveClaudeDebug("%0", content1, cfg) // same as first, but not consecutive — must be written

		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("log file not found: %v", err)
		}
		count := strings.Count(string(data), "=====")
		if count != 6 {
			t.Errorf("expected 6 (3 blocks × 2 markers), got %d:\n%s", count, string(data))
		}
	})

	t.Run("dedup is scoped per pane", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "debug.log")
		cfg := defaultConfig()
		cfg.Debug.Enabled = true
		cfg.Debug.File = logPath

		// %0 writes content1, %1 writes same content1 (different pane — must write),
		// then %0 writes content1 again (duplicate for %0 — must skip).
		saveClaudeDebug("%0", content1, cfg)
		saveClaudeDebug("%1", content1, cfg) // different pane — written
		saveClaudeDebug("%0", content1, cfg) // duplicate for %0 — skipped

		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("log file not found: %v", err)
		}
		body := string(data)
		count := strings.Count(body, "=====")
		if count != 4 {
			t.Errorf("expected 4 (2 blocks × 2 markers), got %d:\n%s", count, body)
		}
		if !strings.Contains(body, "pane=%0") {
			t.Errorf("log does not contain pane=%%0:\n%s", body)
		}
		if !strings.Contains(body, "pane=%1") {
			t.Errorf("log does not contain pane=%%1:\n%s", body)
		}
	})
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
