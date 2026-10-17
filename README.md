# `init`

Bootstraps a fresh machine into aren55555's terminal/dev setup, interactively. Supports macOS, Arch (Omarchy), and Ubuntu (native or WSL).

## Usage

On a fresh machine:

```sh
curl -fsSL https://raw.githubusercontent.com/aren55555/init/main/bootstrap.sh | sh
```

Or from a checkout: `./bootstrap.sh`. It clones to `~/dev/init` (override with `INIT_DIR`), installs prerequisites and Go, then starts an installer that:

1. confirms the detected platform,
2. lets you pick from the tools/apps in `packages.toml` for this platform,
3. installs Oh My Zsh,
4. symlinks dotfiles into `$HOME`, backing up anything it replaces to `~/.init-backup/<timestamp>/`,
5. makes zsh the login shell.

Every step skips work that's already done, so re-running is safe. Flags: `--dry-run` (print without changing anything), `--yes` (no prompts).

## Customizing

- **Add a tool or app:** add a `[[tool]]` or `[[app]]` entry to `packages.toml` with the package name for each manager it should install from.
- **Shell config:** shared settings go in `dotfiles/zsh/.zshrc`, per-platform settings in `dotfiles/zsh/os/<platform>.zsh`.
- **More dotfiles:** add the file under `dotfiles/` and a `[[link]]` entry to `dotfiles/links.toml`.
