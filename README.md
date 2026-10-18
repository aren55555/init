# `init`

Bootstraps a fresh machine into aren55555's terminal/dev setup, interactively. Supports macOS, Arch (Omarchy), and Ubuntu (native or WSL).

## Install

On a fresh machine:

```sh
curl -fsSL https://init.aren55555.com | sh
```

Pass flags after `sh -s --`, e.g. to preview without changing anything:

```sh
curl -fsSL https://init.aren55555.com | sh -s -- --dry-run
```

Or from a checkout: `./bootstrap.sh`. `init.aren55555.com` redirects to [`bootstrap.sh`](bootstrap.sh) on `main`. The script clones this repo to `~/dev/init` (override with `INIT_DIR`), installs prerequisites and Go, then starts an installer that:

1. confirms the detected platform,
2. lets you pick from the tools, apps and fonts in `packages.toml` for this platform (all pre-selected),
3. installs Oh My Zsh,
4. symlinks dotfiles into `$HOME`, backing up anything it replaces to `~/.init-backup/<timestamp>/`,
5. makes zsh the login shell.

Every step skips work that's already done, so re-running is safe. Flags: `--dry-run` (print without changing anything), `--yes` (no prompts).

## What you get

### On every platform

- **zsh** as the login shell, with **Oh My Zsh** and the **agnoster** theme. agnoster needs a Powerline/Nerd font in the terminal, e.g. Hack Nerd Font.
- **`~/.zshrc`** linked to [`dotfiles/zsh/.zshrc`](dotfiles/zsh/.zshrc), which activates mise and zoxide and sources a per-platform file from `dotfiles/zsh/os/`.
- **Go**, installed globally through mise (or an existing `go` is used).
- **Homebrew** on macOS, installed first, which also installs the Xcode Command Line Tools.

### Packages

From [`packages.toml`](packages.toml). On Arch, "AUR" means it's installed with `yay` from the Arch User Repository; everything else there comes from the official repos.

#### Tools

| | macOS | Arch / Omarchy | Ubuntu | WSL |
|---|:-:|:-:|:-:|:-:|
| git | ✓ | ✓ | ✓ | ✓ |
| gh (GitHub CLI) | ✓ | ✓ | ✓ | ✓ |
| zsh | built in | ✓ | ✓ | ✓ |
| mise | ✓ | ✓ | ✓ | ✓ |
| zoxide | ✓ | ✓ | ✓ | ✓ |
| Claude Code | ✓ | | | |

#### Apps

Never installed on WSL, where they belong on the Windows side.

| | macOS | Arch / Omarchy | Ubuntu | WSL |
|---|:-:|:-:|:-:|:-:|
| Google Chrome | ✓ | AUR | | |
| 1Password | ✓ | AUR | | |
| Ghostty | ✓ | ✓ | | |
| Visual Studio Code | ✓ | AUR | | |
| Signal | ✓ | ✓ | | |
| Fork | ✓ | | | |
| Claude | ✓ | | | |
| GPG Suite | ✓ | | | |
| Rectangle | ✓ | | | |
| Notion | ✓ | | | |
| Notion Calendar | ✓ | | | |
| Mullvad VPN | ✓ | | | |
| Alfred | ✓ | | | |
| Thaw | ✓ | | | |
| Todoist | ✓ | | | |
| Raindrop.io | ✓ | | | |
| Pocket Casts | ✓ | | | |
| Spotify | ✓ | | | |

#### Fonts

Never installed on WSL, where the Windows terminal renders them.

| | macOS | Arch / Omarchy | Ubuntu | WSL |
|---|:-:|:-:|:-:|:-:|
| Amatic SC | ✓ | | | |
| Averia Libre | ✓ | | | |
| Bitstream Vera | ✓ | ✓ | | |
| Bitstream Vera Sans Mono Nerd Font | ✓ | ✓ | | |
| DM Mono | ✓ | | | |
| Hack Nerd Font | ✓ | ✓ | | |
| Intel One Mono | ✓ | | | |
| Pangolin | ✓ | | | |

On macOS everything except mise comes from Homebrew (apps and fonts as casks); mise uses its official install script everywhere. Ubuntu gets no apps or fonts because none of them are in Ubuntu's standard repos.

## Customizing

- **Add a tool, app or font:** add a `[[tool]]`, `[[app]]` or `[[font]]` entry to `packages.toml` with the package name for each manager it should install from, then add it to the table above (a test fails if you forget).
- **Shell config:** shared settings go in `dotfiles/zsh/.zshrc`, per-platform settings in `dotfiles/zsh/os/<platform>.zsh`.
- **More dotfiles:** add the file under `dotfiles/` and a `[[link]]` entry to `dotfiles/links.toml`.

## Development

Dev tools come from [`mise.toml`](mise.toml): run `mise trust && mise install` once. Then `just` lists the recipes, including `just vm try`, which bootstraps a fresh macOS VM and opens a shell in it. See [CLAUDE.md](CLAUDE.md) for details.
