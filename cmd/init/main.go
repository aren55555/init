// Command init interactively configures this machine. bootstrap.sh runs it from the repo root.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/aren55555/init/internal/install"
	"github.com/aren55555/init/internal/manifest"
	"github.com/aren55555/init/internal/platform"
	"github.com/aren55555/init/internal/steps"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "print what would run without changing anything")
	yes := flag.Bool("yes", false, "accept every default without prompting")
	flag.Parse()
	if err := run(*dryRun, *yes); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}
}

func run(dryRun, yes bool) error {
	p, err := platform.Detect()
	if err != nil {
		return err
	}
	repo, err := os.Getwd()
	if err != nil {
		return err
	}
	m, err := manifest.Load(filepath.Join(repo, "packages.toml"))
	if err != nil {
		return fmt.Errorf("%w (run from the repo root)", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	r := &install.Runner{DryRun: dryRun, Out: os.Stdout}

	confirm := func(title string) (bool, error) {
		if yes {
			return true, nil
		}
		ok := true
		err := huh.NewConfirm().Title(title).Value(&ok).Run()
		return ok, err
	}

	if ok, err := confirm(fmt.Sprintf("Configure this machine as %s?", p)); err != nil || !ok {
		return err
	}

	items := m.Resolve(p)
	if !yes {
		if items, err = choose(items); err != nil {
			return err
		}
	}
	failed := steps.Packages(r, items)

	for _, s := range []struct {
		name, question string
		fn             func() error
	}{
		{"Oh My Zsh", "Install Oh My Zsh?", func() error { return steps.OhMyZsh(r, home) }},
		{"dotfiles", "Link dotfiles into " + home + "?", func() error { return steps.Dotfiles(r, repo, home, time.Now()) }},
		{"login shell", "Make zsh your login shell?", func() error { return steps.LoginShell(r) }},
	} {
		ok, err := confirm(s.question)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := s.fn(); err != nil {
			r.Logf("✗ %s: %v", s.name, err)
			failed = append(failed, s.name)
		}
	}

	if len(failed) > 0 {
		return errors.New("failed: " + strings.Join(failed, ", "))
	}
	r.Logf("Done. Open a new terminal to load the shell config.")
	return nil
}

// choose asks which items to install, with everything selected by default.
func choose(items []manifest.Item) ([]manifest.Item, error) {
	opts := make([]huh.Option[int], len(items))
	for i, it := range items {
		opts[i] = huh.NewOption(fmt.Sprintf("%s (%s via %s)", it.Name, it.Kind, it.Method), i).Selected(true)
	}
	var picked []int
	err := huh.NewMultiSelect[int]().
		Title("What should be installed?").
		Options(opts...).
		Value(&picked).
		Run()
	if err != nil {
		return nil, err
	}
	chosen := make([]manifest.Item, len(picked))
	for i, idx := range picked {
		chosen[i] = items[idx]
	}
	return chosen, nil
}
