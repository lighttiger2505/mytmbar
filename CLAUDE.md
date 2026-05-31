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

1. **Claude 実行中の判定**  
   `isClaude(title, cmd)`（title の先頭 rune がブレイルや spinner 文字 → Claude の title パターン、かつ cmd が `"claude"` か版番号）または `hasAgentChild(pid)`（子プロセスに `claude` が存在）が true なら、`capturePane` で pane テキストを取得し `claudeWindowStatus` で状態を解析。  
   出力例: `🤖Claude[🏃Running]`

2. **通常コマンドの表示**  
   `special_commands`（`~/.config/mytmbar/config.toml`）に含まれないコマンドなら `✅<cmd>` を返す。

3. **Git ステータスの表示**  
   special command（デフォルト: zsh/bash/vim/nvim/tig）なら `gitWindowStatus` を呼ぶ。  
   - 通常リポジトリ: `🌿<repo>`  
   - linked worktree: `🌿<repo> 🌲<branch>`  
   - Git 外: `📁<dirname>`

### ファイル別の役割

- **`claude.go`** — `parseClaudeStatus` が pane テキストを正規表現で解析し `ClaudeState`（Idle/Running/Waiting）と `ClaudeMode`（plan/accept/auto）を返す。正規表現群（`claudeRunningPattern` など）が判定の中核。
- **`window.go`** — `git rev-parse` で toplevel/commonDir/gitDir を比較し worktree を判定。`truncate` はマルチバイト対応（rune 単位）。
- **`config.go`** — TOML 設定ファイル読み込み。読み込み失敗時はエラーを返さずデフォルト値にフォールバックする設計。
- **`tmux.go`** — `tmux capture-pane -t <paneID> -p` のラッパー。

## Conventions

- テーブル駆動テスト（`t.Run` でサブテスト、テスト名は日本語可）
- `.golangci.yml`: errorlint/revive/misspell/modernize を有効化、formatter は gofmt/goimports
- `.claude/settings.json` の PostToolUse フックにより `.go` ファイル編集後に `go build ./...` と `golangci-lint run` が自動実行される
