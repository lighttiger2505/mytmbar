package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// saveClaudeDebug appends a debug block containing the raw captured pane text
// and its parsed Claude status to the configured log file.
// All errors are silently ignored to avoid breaking the status-bar output.
func saveClaudeDebug(paneID, content string, cfg *Config) {
	path := debugLogPath(cfg)
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

	status := parseClaudeStatus(content)
	header := fmt.Sprintf("===== %s pane=%s state=%s mode=%s =====\n",
		time.Now().Format(time.RFC3339),
		paneID,
		status.State.String(),
		status.Mode.String(),
	)
	_, _ = fmt.Fprintf(f, "%s%s\n", header, content)
}
