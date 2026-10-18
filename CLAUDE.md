# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

This repo bootstraps a fresh machine into the owner's preferred terminal/dev setup, through an interactive process. It targets three environments, which should end up configured as alike as possible:

1. macOS
2. Linux: Omarchy (Arch-based) and Ubuntu
3. WSL on Windows (Ubuntu)

Keep platform differences isolated (in `packages.toml` keys and `dotfiles/zsh/os/*.zsh`) so the shared configuration stays common.

## Commands

Dev tools (`just`, `tart`) come from `mise.toml` (`mise trust && mise install`), pinned by URL and sha256 in `mise.lock`, with `locked` and `paranoid` on. In paranoid mode every edit to `mise.toml` needs a fresh `mise trust`. After changing a version, regenerate the lockfile with `mise lock just aqua:cirruslabs/tart`; a bare `mise lock` also writes `~/.config/mise/mise.lock`. Tart's postinstall hook rejects any build not signed by team `9M2P8L4D89` and notarized by Apple. `just` lists the recipes that wrap the commands below.

```sh
go test ./...                          # all tests
go test ./internal/steps -run TestDotfiles
go run ./cmd/init --dry-run --yes      # see what would run on this machine, with no prompts or changes
sh -n bootstrap.sh && zsh -n dotfiles/zsh/.zshrc   # syntax-check the shell files
```

Don't run the installer on the dev machine without `--dry-run`: it replaces `~/.zshrc` and installs packages.

End-to-end test on a fresh macOS VM (needs Tart; the first run downloads a ~24 GB image into the `mac-base` VM):

```sh
go run ./cmd/vmtest           # bootstrap the local checkout (shared into the VM read-only), run it twice, check the end state
go run ./cmd/vmtest -remote   # same, but via the published `curl -fsSL https://init.aren55555.com | sh` one-liner
go run ./cmd/vmtest -keep     # keep the VM afterwards for poking around
```

Logs go to `.logs/` (gitignored): `<vm>.log` for each test run, and `<vm>-shell-<time>.log` recording each `just ssh shell` session. The first bootstrap run inside the VM takes about 5–7 minutes, mostly Homebrew and cask installs.

## Architecture

- `bootstrap.sh`: POSIX sh, the only code that runs on a bare machine. It installs prerequisites (Homebrew / git+curl), gets Go (existing `go`, or via mise), clones the repo if needed, then `exec`s `go run ./cmd/init`. All logic sits inside `main` so `curl | sh` can reattach stdin to `/dev/tty` for the TUI.
- `cmd/init`: the interactive installer (charmbracelet/huh). `--yes` skips every prompt; `--dry-run` makes `install.Runner` print commands instead of running them.
- `internal/platform`: detects macos / arch / ubuntu, plus a WSL flag, from GOOS, `/etc/os-release` (`ID`, `ID_LIKE`) and `/proc/version`.
- `internal/manifest`: `packages.toml` → `[]Item` for a platform. Each entry has per-manager keys. Method priority: macOS `cask`>`brew`, Arch `pacman`>`aur`, Ubuntu `apt`, then `script` everywhere. `[[app]]` and `[[font]]` entries are skipped on WSL.
- `internal/install`: per-method "installed?" checks and install commands. Read-only checks run even in dry-run mode.
- `internal/steps`: packages, Oh My Zsh, dotfile symlinks (from `dotfiles/links.toml`), login shell. Each step must be idempotent.
- `dotfiles/zsh/.zshrc` finds its own directory through the `~/.zshrc` symlink (`${(%):-%x}:A:h`) and sources `os/<platform>.zsh` before Oh My Zsh, mise and zoxide.

## Constraints

- No bun/node in the bootstrapper: only sh and Go.
- No secrets on the filesystem. Dotfiles must not contain or template in credentials.
