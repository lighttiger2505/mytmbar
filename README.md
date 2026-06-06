# mytmbar
Minimal tmux status bar. built for personal use.

## Configuration

mytmbar reads an optional TOML config file from:

```
$XDG_CONFIG_HOME/mytmbar/config.toml   # if XDG_CONFIG_HOME is set
~/.config/mytmbar/config.toml          # default
```

All keys are optional; missing keys fall back to built-in defaults.
See [`config.toml.example`](config.toml.example) for the full list of configurable options with their default values.
