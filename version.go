package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Injected at build time via -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "unknown"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "print version information",
		Action: runVersion,
	}
}

func runVersion(_ *cli.Context) error {
	fmt.Println(versionString(version, commit))
	return nil
}

func versionString(v, c string) string {
	return fmt.Sprintf("mytmbar %s (%s)", v, c)
}
