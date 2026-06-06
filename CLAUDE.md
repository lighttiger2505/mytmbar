# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build    # go build .
make test     # go build + go test -v ./...
make lint     # golangci-lint run
make install  # go install .
```

単一テストの実行:

```bash
go test -run TestParseClaudeStatus -v
```

ツールバージョンは `mise.toml` で管理（go 1.26.3 / golangci-lint 2.12.2）。

## Architecture

`mytmbar` は tmux の window status 文字列を生成する単一パッケージ（`package main`）の CLI。tmux から各 pane の情報（`--path/--cmd/--pid/--title/--pane-id`）をフラグとして受け取り、ステータス文字列を stdout に出力する。

### 判定フロー（`generateContent` in `main.go`）

出力は常に `<ディレクトリ情報> [コマンド枠]` の形。ディレクトリ情報を先に算出し、その後コマンド枠を出し分ける。

**ディレクトリ情報**（`gitWindowStatus` in `window.go`）:
- 通常リポジトリ: `🌿<repo>`
- linked worktree: `🌿<repo> 🌲<branch>`
- Git 外: `📁<dirname>`

**コマンド枠**（3 パターン）:

1. **Claude 実行中**  
   `isClaude(title, cmd)` または `hasAgentChild(pid)` が true なら、`capturePane` で pane テキストを取得し `claudeWindowStatus` で状態を解析してコマンド枠に入れる。`🚀claude` は出さない。  
   出力例: `🌿myrepo │ 🤖Claude[🏃Running]`

2. **シェル**（`shellCommands` in `window.go`: zsh/bash/sh/fish/tcsh/csh/ksh/dash/nu）  
   コマンド枠なし。ディレクトリのみ（セパレータなし）。  
   出力例: `🌿myrepo`

3. **その他のコマンド**（vim/nvim/tig/npm/go など）  
   `│ 🚀<cmd>` をコマンド枠として付与。  
   出力例: `🌿myrepo │ 🚀vim`、`📁tmp │ 🚀go`

### ファイル別の役割

- **`claude.go`** — `parseClaudeStatus` が pane テキストを正規表現で解析し `ClaudeState`（Idle/Running/Waiting）と `ClaudeMode`（plan/accept/auto）を返す。正規表現群（`claudeRunningPattern` など）が判定の中核。
- **`window.go`** — `git rev-parse` で toplevel/commonDir/gitDir を比較し worktree を判定。`truncate` はマルチバイト対応（rune 単位）。シェルコマンドのリスト（`shellCommands`）と判定関数（`isShellCommand`）もここ。
- **`tmux.go`** — `tmux capture-pane -t <paneID> -p` のラッパー。

## Conventions

- テーブル駆動テスト（`t.Run` でサブテスト、テスト名は日本語可）
- `.golangci.yml`: errorlint/revive/misspell/modernize を有効化、formatter は gofmt/goimports
- `.claude/settings.json` の PostToolUse フックにより `.go` ファイル編集後に `go build ./...` と `golangci-lint run` が自動実行される
