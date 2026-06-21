package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/urfave/cli/v2"
)

type windowFlags struct {
	dirpath string
	cmd     string
	panePID int
	title   string
	paneID  string
}

func windowCommand() *cli.Command {
	return &cli.Command{
		Name:  "window",
		Usage: "generate tmux window status string",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Usage: "tmux pane current path"},
			&cli.StringFlag{Name: "cmd", Usage: "tmux window current command"},
			&cli.IntFlag{Name: "pid", Usage: "tmux pane pid"},
			&cli.StringFlag{Name: "title", Usage: "tmux pane title"},
			&cli.StringFlag{Name: "pane-id", Usage: "tmux pane id"},
		},
		Action: runWindow,
	}
}

func runWindow(c *cli.Context) error {
	flags := &windowFlags{
		dirpath: c.String("path"),
		cmd:     c.String("cmd"),
		panePID: c.Int("pid"),
		title:   c.String("title"),
		paneID:  c.String("pane-id"),
	}

	cfg := LoadConfig()
	content, err := generateContent(flags, cfg)
	if err != nil {
		return err
	}
	fmt.Printf(" %s \n", content)
	return nil
}

func generateContent(flags *windowFlags, cfg *Config) (string, error) {
	status, err := gitWindowStatus(flags.dirpath, cfg)
	if err != nil {
		return "", err
	}

	// Claude takes the command slot; never show the command icon for claude.
	if isClaude(flags.title, flags.cmd) || (flags.panePID > 0 && hasAgentChild(flags.panePID)) {
		label := "Claude"
		if cfg.Display.Short {
			label = "CC"
		}
		claudePart := cfg.Icons.Claude + label
		if content, err := capturePane(flags.paneID); err == nil && content != "" {
			if cfg.Debug.Enabled {
				saveClaudeDebug(flags.paneID, content, cfg)
			}
			claudePart = claudeWindowStatus(content, cfg)
		}
		return fmt.Sprintf("%s %s %s", status, cfg.Icons.Separator, claudePart), nil
	}

	// Shell commands carry no useful name; show directory only.
	if isShellCommand(flags.cmd) {
		return status, nil
	}

	return fmt.Sprintf("%s %s %s%s", status, cfg.Icons.Separator, cfg.Icons.Command, truncate(flags.cmd, cfg.Lengths.Command)), nil
}

func hasAgentChild(pid int) bool {
	children, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	procs, err := children.Children()
	if err != nil {
		return false
	}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if name == "claude" || mayBeClaudeProcess(name) {
			return true
		}
	}
	return false
}

var shellCommands = []string{"zsh", "bash", "sh", "fish", "tcsh", "csh", "ksh", "dash", "nu"}

func isShellCommand(cmd string) bool {
	return slices.Contains(shellCommands, cmd)
}

func gitWindowStatus(dirpath string, cfg *Config) (string, error) {
	icons := cfg.Icons
	lengths := cfg.Lengths

	topLevel, err := runGit(dirpath, "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Sprintf("%s%s", icons.Directory, truncate(filepath.Base(dirpath), lengths.Directory)), nil
	}

	commonDir, err := runGit(dirpath, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return fmt.Sprintf("%s%s", icons.Repo, truncate(filepath.Base(topLevel), lengths.Directory)), nil
	}

	gitDir, err := runGit(dirpath, "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return fmt.Sprintf("%s%s", icons.Repo, truncate(filepath.Base(topLevel), lengths.Directory)), nil
	}

	if commonDir != gitDir {
		// linked worktree: gitDir is under <commonDir>/worktrees/<name>
		repoName := filepath.Base(filepath.Dir(commonDir))
		branch, err := runGit(dirpath, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return fmt.Sprintf("%s%s", icons.Repo, truncate(repoName, lengths.Directory)), nil
		}
		return fmt.Sprintf("%s%s %s%s", icons.Repo, truncate(repoName, lengths.Directory), icons.WorktreeBranch, truncate(branch, lengths.Branch)), nil
	}

	return fmt.Sprintf("%s%s", icons.Repo, truncate(filepath.Base(topLevel), lengths.Directory)), nil
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
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
