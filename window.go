package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func isSpecialCommand(cmd string, specialCmds []string) bool {
	return slices.Contains(specialCmds, cmd)
}

func gitWindowStatus(dirpath string) (string, error) {
	topLevel, err := runGit(dirpath, "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Sprintf("📁%s", filepath.Base(dirpath)), nil
	}

	commonDir, err := runGit(dirpath, "rev-parse", "--git-common-dir")
	if err != nil {
		return fmt.Sprintf("🌿%s", filepath.Base(topLevel)), nil
	}

	gitDir, err := runGit(dirpath, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Sprintf("🌿%s", filepath.Base(topLevel)), nil
	}

	if commonDir != gitDir {
		// worktree: commonDir points to main repo's .git, gitDir points to worktree's .git
		repoName := filepath.Base(filepath.Dir(commonDir))
		worktreeName := filepath.Base(topLevel)
		return fmt.Sprintf("🌿%s 🌲%s", repoName, truncate(worktreeName, 16)), nil
	}

	return fmt.Sprintf("🌿%s", filepath.Base(topLevel)), nil
}

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) >= maxLen {
		return string(r[:maxLen-1]) + "…"
	}
	return s
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
