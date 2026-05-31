package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var defaultSpecialCommands = []string{"zsh", "bash"}

type Config struct {
	SpecialCommands []string `toml:"special_commands"`
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig(), nil
	}

	path := filepath.Join(home, ".config", "mytmbar", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig(), nil
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), nil
	}

	if len(cfg.SpecialCommands) == 0 {
		cfg.SpecialCommands = defaultSpecialCommands
	}
	return &cfg, nil
}

func defaultConfig() *Config {
	return &Config{SpecialCommands: defaultSpecialCommands}
}
