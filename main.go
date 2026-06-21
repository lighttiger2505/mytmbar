package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mytmbar",
		Usage: "tmux window status generator",
		Commands: []*cli.Command{
			windowCommand(),
			configCommand(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		appendErrorLog(LoadConfig(), err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
