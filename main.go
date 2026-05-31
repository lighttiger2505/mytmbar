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
			&cli.StringFlag{Name: "pane-id", Usage: "tmux pane id"},
		},
		Action: run,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type Flags struct {
	dirpath string
	cmd     string
	panePID int
	title   string
	paneID  string
}

func run(c *cli.Context) error {
	flags := &Flags{
		dirpath: c.String("path"),
		cmd:     c.String("cmd"),
		panePID: c.Int("pid"),
		title:   c.String("title"),
		paneID:  c.String("pane-id"),
	}

	content, err := generateContent(flags)
	if err != nil {
		return err
	}
	fmt.Printf(" %s \n", content)
	return nil
}

func generateContent(flags *Flags) (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	status, err := gitWindowStatus(flags.dirpath)
	if err != nil {
		return "", err
	}

	// Claude takes the command slot; never show "✅claude".
	if isClaude(flags.title, flags.cmd) || (flags.panePID > 0 && hasAgentChild(flags.panePID)) {
		claudePart := "🤖Claude"
		if content, err := capturePane(flags.paneID); err == nil && content != "" {
			claudePart = claudeWindowStatus(content)
		}
		return fmt.Sprintf("%s %s", status, claudePart), nil
	}

	// Shell commands carry no useful name; show directory only.
	if isSpecialCommand(flags.cmd, cfg.SpecialCommands) {
		return status, nil
	}

	return fmt.Sprintf("%s ✅%s", status, flags.cmd), nil
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
