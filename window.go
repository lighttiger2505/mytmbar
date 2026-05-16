package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func isSpecialCommand(cmd string, specialCmds []string) bool {
	for _, s := range specialCmds {
		if cmd == s {
			return true
		}
	}
	return false
}

func gitWindowStatus(dirpath string) (string, error) {
	topLevel, err := runGit(dirpath, "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Sprintf("📁 %s", filepath.Base(dirpath)), nil
	}

	commonDir, err := runGit(dirpath, "rev-parse", "--git-common-dir")
	if err != nil {
		return fmt.Sprintf("🌿 %s", filepath.Base(topLevel)), nil
	}

	gitDir, err := runGit(dirpath, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Sprintf("🌿 %s", filepath.Base(topLevel)), nil
	}

	if commonDir != gitDir {
		// worktree: commonDir points to main repo's .git, gitDir points to worktree's .git
		repoName := filepath.Base(filepath.Dir(commonDir))
		worktreeName := filepath.Base(topLevel)
		return fmt.Sprintf("🌿 %s 🌲 %s", repoName, worktreeName), nil
	}

	return fmt.Sprintf("🌿 %s", filepath.Base(topLevel)), nil
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
