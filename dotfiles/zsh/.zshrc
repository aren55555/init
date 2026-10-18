# Managed by github.com/aren55555/init: ~/.zshrc is a symlink to this file.

# Directory holding this file, resolved through the ~/.zshrc symlink.
INIT_ZSH="${${(%):-%x}:A:h}"

export PATH="$HOME/.local/bin:$PATH"

# Platform-specific setup runs first so the tools below are on PATH.
case "$OSTYPE" in
  darwin*) _init_os=macos ;;
  linux*) _init_os=$(. /etc/os-release && print -r -- "$ID") ;;
esac
[[ -r "$INIT_ZSH/os/$_init_os.zsh" ]] && source "$INIT_ZSH/os/$_init_os.zsh"
unset _init_os

export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME="agnoster"  # needs a Powerline/Nerd font in the terminal (e.g. Hack Nerd Font)
plugins=(git)
[[ -r "$ZSH/oh-my-zsh.sh" ]] && source "$ZSH/oh-my-zsh.sh"

(( $+commands[mise] )) && eval "$(mise activate zsh)"
(( $+commands[zoxide] )) && eval "$(zoxide init zsh)"
