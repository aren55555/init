// Package manifest loads packages.toml and resolves which entries apply to a platform.
package manifest

import (
	"github.com/BurntSushi/toml"

	"github.com/aren55555/init/internal/platform"
)

// Method is how an item gets installed.
type Method string

const (
	Brew   Method = "brew"
	Cask   Method = "cask"
	Pacman Method = "pacman"
	AUR    Method = "aur"
	Apt    Method = "apt"
	Script Method = "script"
)

// Entry is one [[tool]] or [[app]] table; each field names the package for that method.
type Entry struct {
	Name   string `toml:"name"`
	Brew   string `toml:"brew"`
	Cask   string `toml:"cask"`
	Pacman string `toml:"pacman"`
	AUR    string `toml:"aur"`
	Apt    string `toml:"apt"`
	Script string `toml:"script"`
	Check  string `toml:"check"`
}

type Manifest struct {
	Tools []Entry `toml:"tool"`
	Apps  []Entry `toml:"app"`
}

// Item is an entry resolved for one platform. For Script items Pkg holds the
// shell command and Check the command whose presence means it's installed.
type Item struct {
	Name   string
	App    bool
	Method Method
	Pkg    string
	Check  string
}

func Load(path string) (*Manifest, error) {
	var m Manifest
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Resolve returns the items installable on p, in manifest order. Apps are
// skipped on WSL, where GUI apps belong on the Windows side.
func (m *Manifest) Resolve(p platform.Platform) []Item {
	var items []Item
	for _, e := range m.Tools {
		if it, ok := resolve(e, p); ok {
			items = append(items, it)
		}
	}
	if !p.WSL {
		for _, e := range m.Apps {
			if it, ok := resolve(e, p); ok {
				it.App = true
				items = append(items, it)
			}
		}
	}
	return items
}

func resolve(e Entry, p platform.Platform) (Item, bool) {
	type candidate struct {
		method Method
		pkg    string
	}
	var candidates []candidate
	switch p.OS {
	case platform.MacOS:
		candidates = []candidate{{Cask, e.Cask}, {Brew, e.Brew}}
	case platform.Arch:
		candidates = []candidate{{Pacman, e.Pacman}, {AUR, e.AUR}}
	case platform.Ubuntu:
		candidates = []candidate{{Apt, e.Apt}}
	}
	candidates = append(candidates, candidate{Script, e.Script})
	for _, c := range candidates {
		if c.pkg == "" {
			continue
		}
		check := e.Check
		if check == "" {
			check = e.Name
		}
		return Item{Name: e.Name, Method: c.method, Pkg: c.pkg, Check: check}, true
	}
	return Item{}, false
}
