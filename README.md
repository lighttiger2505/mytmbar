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

### Editing the config file

```
mytmbar config edit
```

Opens the config file in `$EDITOR` (falls back to `vi` when `$EDITOR` is unset).
If the config directory does not exist yet, it is created automatically.
The file itself is created by the editor when you save.
