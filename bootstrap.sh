#!/bin/sh
# Installs just enough to run the Go installer (cmd/init), then hands off to it.
#
#   curl -fsSL https://raw.githubusercontent.com/aren55555/init/main/bootstrap.sh | sh
#   ./bootstrap.sh [--dry-run] [--yes]
#
# Flags are passed to the Go installer. Missing prerequisites (Homebrew on
# macOS, git and curl on Linux, Go) are installed even with --dry-run.
#
# Everything is inside main so that, under `curl | sh`, the whole script has
# been read before stdin is reattached to the terminal.
set -eu

REPO_URL="${INIT_REPO_URL:-https://github.com/aren55555/init.git}"
INIT_DIR="${INIT_DIR:-$HOME/dev/init}"

say() { printf '==> %s\n' "$*"; }
die() { printf 'bootstrap: %s\n' "$*" >&2; exit 1; }
has() { command -v "$1" >/dev/null 2>&1; }

brew_shellenv() {
	for brew in /opt/homebrew/bin/brew /usr/local/bin/brew; do
		if [ -x "$brew" ]; then
			eval "$("$brew" shellenv)"
			return
		fi
	done
}

main() {
	if [ ! -t 0 ] && (: </dev/tty) 2>/dev/null; then
		exec </dev/tty
	fi
	if [ "$(id -u)" -eq 0 ]; then sudo=""; else sudo="sudo"; fi

	case "$(uname -s)" in
	Darwin) os=macos ;;
	Linux)
		[ -r /etc/os-release ] || die "cannot read /etc/os-release"
		. /etc/os-release
		case " ${ID:-} ${ID_LIKE:-} " in
		*" arch "*) os=arch ;;
		*" ubuntu "*) os=ubuntu ;;
		*) die "unsupported Linux distribution: ${ID:-unknown}" ;;
		esac
		;;
	*) die "unsupported OS: $(uname -s)" ;;
	esac
	say "Detected $os"

	case $os in
	macos)
		# Homebrew may be installed but not on PATH in a shell that hasn't loaded ~/.zshrc.
		brew_shellenv
		if ! has brew; then
			say "Installing Homebrew (this also installs the Xcode Command Line Tools)"
			/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
			brew_shellenv
		fi
		;;
	arch)
		if ! has git || ! has curl; then
			say "Installing git, curl and base-devel"
			$sudo pacman -Syu --needed --noconfirm git curl base-devel
		fi
		;;
	ubuntu)
		if ! has git || ! has curl; then
			say "Installing git, curl and build-essential"
			$sudo apt-get update
			$sudo env DEBIAN_FRONTEND=noninteractive apt-get install -y git curl ca-certificates build-essential
		fi
		;;
	esac

	script_dir=$(cd "$(dirname "$0")" 2>/dev/null && pwd || true)
	if [ -n "$script_dir" ] && [ -f "$script_dir/packages.toml" ]; then
		INIT_DIR=$script_dir
	elif [ ! -d "$INIT_DIR/.git" ]; then
		say "Cloning $REPO_URL into $INIT_DIR"
		mkdir -p "$(dirname "$INIT_DIR")"
		git clone "$REPO_URL" "$INIT_DIR"
	elif [ -z "$(git -C "$INIT_DIR" status --porcelain 2>/dev/null)" ]; then
		# A clone from an earlier run would otherwise keep running old code.
		say "Updating $INIT_DIR"
		git -C "$INIT_DIR" pull --ff-only || say "Warning: could not update $INIT_DIR; running it as is"
	fi
	cd "$INIT_DIR"

	# Show which commit runs, and warn if GitHub's main differs: an existing
	# clone is never updated, so a re-run can be on old code.
	head=$(git rev-parse HEAD 2>/dev/null || true)
	say "Running $INIT_DIR at commit $(printf '%.12s' "${head:-unknown}")"
	latest=$(git ls-remote "$REPO_URL" refs/heads/main 2>/dev/null | cut -f1 || true)
	if [ -n "$head" ] && [ -n "$latest" ] && [ "$head" != "$latest" ]; then
		say "Warning: main on GitHub is at $(printf '%.12s' "$latest"); to update: git -C \"$INIT_DIR\" pull"
	fi

	if has go; then
		set -- go run ./cmd/init "$@"
	else
		PATH="$HOME/.local/bin:$PATH"
		if ! has mise; then
			say "Installing mise"
			curl -fsSL https://mise.run | sh
		fi
		# Run mise from $HOME so the repo's mise.toml (dev tools only) isn't
		# loaded: it would ask to be trusted and install tools we don't need here.
		if ! mise -C "$HOME" which go >/dev/null 2>&1; then
			say "Installing Go with mise"
			mise -C "$HOME" use -g go@latest
		fi
		set -- "$(mise -C "$HOME" which go)" run ./cmd/init "$@"
	fi
	say "Starting installer"
	exec "$@"
}

main "$@"
