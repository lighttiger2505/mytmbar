package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// appendErrorLog appends a timestamped error entry to the configured error log.
// All errors during writing are silently ignored to avoid masking the original
// error. Does nothing when no path can be determined.
func appendErrorLog(cfg *Config, e error) {
	path := errorLogPath(cfg)
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "%s %v\n", time.Now().Format(time.RFC3339), e)
}
