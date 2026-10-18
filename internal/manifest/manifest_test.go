package manifest

import (
	"reflect"
	"testing"

	"github.com/aren55555/init/internal/platform"
)

var testManifest = &Manifest{
	Tools: []Entry{
		{Name: "gh", Brew: "gh", Pacman: "github-cli", Apt: "gh"},
		{Name: "zsh", Pacman: "zsh", Apt: "zsh"},
		{Name: "mise", Script: "curl -fsSL https://mise.run | sh"},
	},
	Apps: []Entry{
		{Name: "Google Chrome", Cask: "google-chrome", AUR: "google-chrome"},
	},
	Fonts: []Entry{
		{Name: "Hack Nerd Font", Cask: "font-hack-nerd-font", Pacman: "ttf-hack-nerd"},
	},
}

func TestResolve(t *testing.T) {
	mise := Item{Name: "mise", Kind: Tool, Method: Script, Pkg: "curl -fsSL https://mise.run | sh", Check: "mise"}
	tests := []struct {
		name string
		p    platform.Platform
		want []Item
	}{
		{"macos", platform.Platform{OS: platform.MacOS}, []Item{
			{Name: "gh", Kind: Tool, Method: Brew, Pkg: "gh", Check: "gh"},
			mise,
			{Name: "Google Chrome", Kind: App, Method: Cask, Pkg: "google-chrome", Check: "Google Chrome"},
			{Name: "Hack Nerd Font", Kind: Font, Method: Cask, Pkg: "font-hack-nerd-font", Check: "Hack Nerd Font"},
		}},
		{"arch", platform.Platform{OS: platform.Arch}, []Item{
			{Name: "gh", Kind: Tool, Method: Pacman, Pkg: "github-cli", Check: "gh"},
			{Name: "zsh", Kind: Tool, Method: Pacman, Pkg: "zsh", Check: "zsh"},
			mise,
			{Name: "Google Chrome", Kind: App, Method: AUR, Pkg: "google-chrome", Check: "Google Chrome"},
			{Name: "Hack Nerd Font", Kind: Font, Method: Pacman, Pkg: "ttf-hack-nerd", Check: "Hack Nerd Font"},
		}},
		{"wsl skips apps and fonts", platform.Platform{OS: platform.Ubuntu, WSL: true}, []Item{
			{Name: "gh", Kind: Tool, Method: Apt, Pkg: "gh", Check: "gh"},
			{Name: "zsh", Kind: Tool, Method: Apt, Pkg: "zsh", Check: "zsh"},
			mise,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testManifest.Resolve(tt.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestLoadRepoManifest(t *testing.T) {
	m, err := Load("../../packages.toml")
	if err != nil {
		t.Fatal(err)
	}
	for _, os := range []platform.OS{platform.MacOS, platform.Arch, platform.Ubuntu} {
		if len(m.Resolve(platform.Platform{OS: os})) == 0 {
			t.Errorf("no items resolved for %s", os)
		}
	}
}
