package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestErrorLogPath(t *testing.T) {
	home := t.TempDir()

	tests := []struct {
		name       string
		configFile string
		xdgState   string
		wantSuffix string
	}{
		{
			name:       "config File override is returned as-is",
			configFile: "/custom/path/error.log",
			wantSuffix: "/custom/path/error.log",
		},
		{
			name:       "XDG_STATE_HOME is used when File is empty",
			xdgState:   filepath.Join(home, "state"),
			wantSuffix: filepath.Join(home, "state", "mytmbar", "error.log"),
		},
		{
			name:       "fallback to ~/.local/state when XDG_STATE_HOME is unset",
			wantSuffix: filepath.Join(".local", "state", "mytmbar", "error.log"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", tt.xdgState)
			cfg := defaultConfig()
			cfg.Error.File = tt.configFile

			got := errorLogPath(cfg)

			if tt.configFile != "" {
				if got != tt.configFile {
					t.Errorf("errorLogPath() = %q, want %q", got, tt.configFile)
				}
				return
			}
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("errorLogPath() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestAppendErrorLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "error.log")
	cfg := defaultConfig()
	cfg.Error.File = logPath

	appendErrorLog(cfg, errors.New("something went wrong"))

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file was not created: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "something went wrong") {
		t.Errorf("log does not contain error message:\n%s", body)
	}
}

func TestAppendErrorLog_Append(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "error.log")
	cfg := defaultConfig()
	cfg.Error.File = logPath

	appendErrorLog(cfg, errors.New("first error"))
	appendErrorLog(cfg, errors.New("second error"))

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file was not created: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "first error") {
		t.Errorf("first error missing from log:\n%s", body)
	}
	if !strings.Contains(body, "second error") {
		t.Errorf("second error missing from log:\n%s", body)
	}
}
