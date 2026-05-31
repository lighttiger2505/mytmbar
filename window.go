package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var shellCommands = []string{"zsh", "bash", "sh", "fish", "tcsh", "csh", "ksh", "dash", "nu"}

func isShellCommand(cmd string) bool {
	return slices.Contains(shellCommands, cmd)
}

func gitWindowStatus(dirpath string) (string, error) {
	topLevel, err := runGit(dirpath, "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Sprintf("📁%s", filepath.Base(dirpath)), nil
	}

	commonDir, err := runGit(dirpath, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return fmt.Sprintf("🌿%s", filepath.Base(topLevel)), nil
	}

	gitDir, err := runGit(dirpath, "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return fmt.Sprintf("🌿%s", filepath.Base(topLevel)), nil
	}

	if commonDir != gitDir {
		// linked worktree: gitDir is under <commonDir>/worktrees/<name>
		repoName := filepath.Base(filepath.Dir(commonDir))
		branch, err := runGit(dirpath, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return fmt.Sprintf("🌿%s", repoName), nil
		}
		return fmt.Sprintf("🌿%s 🌲%s", repoName, truncate(branch, 16)), nil
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
