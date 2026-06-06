package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	// lengths
	if cfg.Lengths.Directory != 20 {
		t.Errorf("Lengths.Directory = %d, want 20", cfg.Lengths.Directory)
	}
	if cfg.Lengths.Branch != 16 {
		t.Errorf("Lengths.Branch = %d, want 16", cfg.Lengths.Branch)
	}
	if cfg.Lengths.Command != 16 {
		t.Errorf("Lengths.Command = %d, want 16", cfg.Lengths.Command)
	}

	// directory icons
	if cfg.Icons.Repo != "🌿" {
		t.Errorf("Icons.Repo = %q, want 🌿", cfg.Icons.Repo)
	}
	if cfg.Icons.WorktreeBranch != "🌲" {
		t.Errorf("Icons.WorktreeBranch = %q, want 🌲", cfg.Icons.WorktreeBranch)
	}
	if cfg.Icons.Directory != "📁" {
		t.Errorf("Icons.Directory = %q, want 📁", cfg.Icons.Directory)
	}

	// separator
	if cfg.Icons.Separator != "│" {
		t.Errorf("Icons.Separator = %q, want │", cfg.Icons.Separator)
	}

	// command/claude icons
	if cfg.Icons.Command != "🚀" {
		t.Errorf("Icons.Command = %q, want 🚀", cfg.Icons.Command)
	}
	if cfg.Icons.Claude != "🤖" {
		t.Errorf("Icons.Claude = %q, want 🤖", cfg.Icons.Claude)
	}

	// state icons
	if cfg.Icons.StateIdle != "⌛" {
		t.Errorf("Icons.StateIdle = %q, want ⌛", cfg.Icons.StateIdle)
	}
	if cfg.Icons.StateRunning != "🏃" {
		t.Errorf("Icons.StateRunning = %q, want 🏃", cfg.Icons.StateRunning)
	}
	if cfg.Icons.StateWaiting != "🚧" {
		t.Errorf("Icons.StateWaiting = %q, want 🚧", cfg.Icons.StateWaiting)
	}
	if cfg.Icons.StateUnknown != "❓" {
		t.Errorf("Icons.StateUnknown = %q, want ❓", cfg.Icons.StateUnknown)
	}

	// mode icons
	if cfg.Icons.ModePlan != "📋" {
		t.Errorf("Icons.ModePlan = %q, want 📋", cfg.Icons.ModePlan)
	}
	if cfg.Icons.ModeAccept != "✏️" {
		t.Errorf("Icons.ModeAccept = %q, want ✏️", cfg.Icons.ModeAccept)
	}
	if cfg.Icons.ModeAuto != "🔄" {
		t.Errorf("Icons.ModeAuto = %q, want 🔄", cfg.Icons.ModeAuto)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("no file returns defaults", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfg := LoadConfig()

		if cfg.Lengths.Directory != 20 {
			t.Errorf("Lengths.Directory = %d, want 20", cfg.Lengths.Directory)
		}
		if cfg.Icons.Repo != "🌿" {
			t.Errorf("Icons.Repo = %q, want 🌿", cfg.Icons.Repo)
		}
	})

	t.Run("partial TOML overrides only specified keys", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "mytmbar")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := `[lengths]
directory = 30
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg := LoadConfig()

		if cfg.Lengths.Directory != 30 {
			t.Errorf("Lengths.Directory = %d, want 30", cfg.Lengths.Directory)
		}
		// unspecified keys stay at defaults
		if cfg.Lengths.Branch != 16 {
			t.Errorf("Lengths.Branch = %d, want 16 (default)", cfg.Lengths.Branch)
		}
		if cfg.Icons.Command != "🚀" {
			t.Errorf("Icons.Command = %q, want 🚀 (default)", cfg.Icons.Command)
		}
	})

	t.Run("icon override", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "mytmbar")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := `[icons]
command = "🔧"
repo = "📦"
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg := LoadConfig()

		if cfg.Icons.Command != "🔧" {
			t.Errorf("Icons.Command = %q, want 🔧", cfg.Icons.Command)
		}
		if cfg.Icons.Repo != "📦" {
			t.Errorf("Icons.Repo = %q, want 📦", cfg.Icons.Repo)
		}
		// unspecified icon stays default
		if cfg.Icons.Claude != "🤖" {
			t.Errorf("Icons.Claude = %q, want 🤖 (default)", cfg.Icons.Claude)
		}
	})

	t.Run("invalid TOML returns defaults", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "mytmbar")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte("not valid toml ][[["), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg := LoadConfig()

		if cfg.Lengths.Directory != 20 {
			t.Errorf("Lengths.Directory = %d, want 20 (default after parse error)", cfg.Lengths.Directory)
		}
	})
}

func TestConfigPath(t *testing.T) {
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
		got := configPath()
		want := "/tmp/xdg/mytmbar/config.toml"
		if got != want {
			t.Errorf("configPath() = %q, want %q", got, want)
		}
	})

	t.Run("XDG_CONFIG_HOME unset falls back to ~/.config", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		got := configPath()
		if got == "" {
			t.Skip("UserHomeDir not available")
		}
		if filepath.Base(filepath.Dir(got)) != "mytmbar" {
			t.Errorf("configPath() = %q, expected .../.config/mytmbar/config.toml", got)
		}
		if filepath.Base(got) != "config.toml" {
			t.Errorf("configPath() = %q, expected filename config.toml", got)
		}
	})
}
