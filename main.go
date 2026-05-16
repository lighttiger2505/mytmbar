package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mytmbar",
		Usage: "tmux window status generator",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Usage: "tmux pane current path"},
			&cli.StringFlag{Name: "cmd", Usage: "tmux window current command"},
			&cli.Int64Flag{Name: "pid", Usage: "tmux pane pid"},
			&cli.StringFlag{Name: "title", Usage: "tmux pane title"},
		},
		Action: run,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	dirpath := c.String("path")
	cmd := c.String("cmd")
	panePID := c.Int64("pid")
	title := c.String("title")

	// Claude Code detection via title+cmd (tcmux approach)
	if isClaude(title, cmd) {
		fmt.Println(claudeWindowStatus(title))
		return nil
	}

	// Claude Code detection via child process (fallback for shell wrappers)
	if panePID > 0 && hasClaudeChild(int32(panePID)) {
		fmt.Println(claudeWindowStatus(title))
		return nil
	}

	cfg, _ := LoadConfig()

	if !isSpecialCommand(cmd, cfg.SpecialCommands) {
		fmt.Printf("✅ %s\n", cmd)
		return nil
	}

	status, err := gitWindowStatus(dirpath)
	if err != nil {
		return err
	}
	fmt.Println(status)
	return nil
}

func hasClaudeChild(pid int32) bool {
	children, err := process.NewProcess(pid)
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
