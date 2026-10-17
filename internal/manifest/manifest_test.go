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
}

func TestResolve(t *testing.T) {
	mise := Item{Name: "mise", Method: Script, Pkg: "curl -fsSL https://mise.run | sh", Check: "mise"}
	tests := []struct {
		name string
		p    platform.Platform
		want []Item
	}{
		{"macos", platform.Platform{OS: platform.MacOS}, []Item{
			{Name: "gh", Method: Brew, Pkg: "gh", Check: "gh"},
			mise,
			{Name: "Google Chrome", App: true, Method: Cask, Pkg: "google-chrome", Check: "Google Chrome"},
		}},
		{"arch", platform.Platform{OS: platform.Arch}, []Item{
			{Name: "gh", Method: Pacman, Pkg: "github-cli", Check: "gh"},
			{Name: "zsh", Method: Pacman, Pkg: "zsh", Check: "zsh"},
			mise,
			{Name: "Google Chrome", App: true, Method: AUR, Pkg: "google-chrome", Check: "Google Chrome"},
		}},
		{"wsl skips apps", platform.Platform{OS: platform.Ubuntu, WSL: true}, []Item{
			{Name: "gh", Method: Apt, Pkg: "gh", Check: "gh"},
			{Name: "zsh", Method: Apt, Pkg: "zsh", Check: "zsh"},
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
