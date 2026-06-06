package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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

// Config is the top-level configuration structure.
type Config struct {
	Lengths Lengths `toml:"lengths"`
	Icons   Icons   `toml:"icons"`
}

// defaultConfig returns a Config populated with the built-in defaults.
func defaultConfig() *Config {
	return &Config{
		Lengths: Lengths{
			Directory: 20,
			Branch:    16,
			Command:   16,
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
