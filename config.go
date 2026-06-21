package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"
)

// Lengths holds maximum display lengths for each element.
// A value of 0 or less means no truncation.
type Lengths struct {
	Directory int `toml:"directory"`
	Branch    int `toml:"branch"`
	Command   int `toml:"command"`
}

// Icons holds the emoji/string used for each status indicator.
type Icons struct {
	Repo           string `toml:"repo"`
	WorktreeBranch string `toml:"worktree_branch"`
	Directory      string `toml:"directory"`
	Separator      string `toml:"separator"`
	Command        string `toml:"command"`
	Claude         string `toml:"claude"`
	StateIdle      string `toml:"state_idle"`
	StateRunning   string `toml:"state_running"`
	StateWaiting   string `toml:"state_waiting"`
	StateUnknown   string `toml:"state_unknown"`
	ModePlan       string `toml:"mode_plan"`
	ModeAccept     string `toml:"mode_accept"`
	ModeAuto       string `toml:"mode_auto"`
}

// Display holds display-mode toggles.
type Display struct {
	Short bool `toml:"short"`
}

// Debug holds debug-output settings.
type Debug struct {
	Enabled bool     `toml:"enabled"`
	File    string   `toml:"file"`   // override path; empty means default
	States  []string `toml:"states"` // filter by state name; empty means all states
}

// Claude holds Claude-detection settings.
type Claude struct {
	Commands []string `toml:"commands"` // extra process names treated as Claude
}

// Config is the top-level configuration structure.
type Config struct {
	Lengths Lengths `toml:"lengths"`
	Icons   Icons   `toml:"icons"`
	Display Display `toml:"display"`
	Debug   Debug   `toml:"debug"`
	Claude  Claude  `toml:"claude"`
}

// defaultConfig returns a Config populated with the built-in defaults.
func defaultConfig() *Config {
	return &Config{
		Lengths: Lengths{
			Directory: 20,
			Branch:    16,
			Command:   16,
		},
		Display: Display{
			Short: false,
		},
		Debug: Debug{
			Enabled: false,
			File:    "",
		},
		Claude: Claude{
			Commands: []string{"lbox"},
		},
		Icons: Icons{
			Repo:           "🌿",
			WorktreeBranch: "🌲",
			Directory:      "📁",
			Separator:      "│",
			Command:        "🚀",
			Claude:         "🤖",
			StateIdle:      "⌛",
			StateRunning:   "🏃",
			StateWaiting:   "🚧",
			StateUnknown:   "❓",
			ModePlan:       "📋",
			ModeAccept:     "✏️",
			ModeAuto:       "🔄",
		},
	}
}

// LoadConfig reads the config file at the XDG config path.
// Missing file or missing keys fall back to defaultConfig values.
func LoadConfig() *Config {
	cfg := defaultConfig()
	path := configPath()
	if path == "" {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		// File absent or unreadable — use defaults.
		return cfg
	}
	// Unmarshal into the pre-filled default; only present keys are overwritten.
	_ = toml.Unmarshal(data, cfg)
	return cfg
}

func configCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "manage the configuration file",
		Subcommands: []*cli.Command{
			{
				Name:   "edit",
				Usage:  "open the config file in $EDITOR",
				Action: runConfigEdit,
			},
		},
	}
}

func runConfigEdit(_ *cli.Context) error {
	path := configPath()
	if path == "" {
		return fmt.Errorf("cannot determine config path")
	}
	// Create the parent directory; the editor creates the file itself on save.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	editor := resolveEditor()
	// $EDITOR may carry arguments (e.g. "code --wait"); split into argv.
	parts := strings.Fields(editor)
	args := append(parts[1:], path)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// resolveEditor returns $EDITOR, falling back to vi when unset.
func resolveEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

// configPath returns the path to the config file, respecting XDG_CONFIG_HOME.
func configPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "mytmbar", "config.toml")
}

// debugLogPath returns the path to the debug log file.
// When cfg.Debug.File is set it is returned as-is; otherwise the default is
// $XDG_STATE_HOME/mytmbar/debug.log (falling back to ~/.local/state).
func debugLogPath(cfg *Config) string {
	if cfg.Debug.File != "" {
		return cfg.Debug.File
	}
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "mytmbar", "debug.log")
}
