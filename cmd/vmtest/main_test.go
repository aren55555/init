package main

import (
	"testing"

	"github.com/aren55555/init/internal/manifest"
)

func TestInstalledCheck(t *testing.T) {
	tests := []struct {
		item manifest.Item
		want string
	}{
		{manifest.Item{Name: "Claude Code", Method: manifest.Cask, Pkg: "claude-code@latest", Check: "claude"}, `zsh -lic 'brew list --cask claude-code@latest'`},
		{manifest.Item{Name: "gh", Method: manifest.Brew, Pkg: "gh", Check: "gh"}, `zsh -lic 'brew list --formula gh'`},
		{manifest.Item{Name: "mise", Method: manifest.Script, Pkg: "curl -fsSL https://mise.run | sh", Check: "mise"}, `zsh -lic 'command -v mise'`},
	}
	for _, tt := range tests {
		if got := installedCheck(tt.item); got.cmd != tt.want || got.name != tt.item.Name+" installed" {
			t.Errorf("installedCheck(%s) = %+v, want cmd %q", tt.item.Name, got, tt.want)
		}
	}
}
