// Package steps holds the installer's setup steps. Each step is safe to re-run.
package steps

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/aren55555/init/internal/install"
	"github.com/aren55555/init/internal/manifest"
)

// Packages installs each item that isn't already installed and returns the
// names of those that failed and of those skipped because their package
// manager wasn't found.
func Packages(r *install.Runner, items []manifest.Item) (failed, skipped []string) {
	for _, it := range items {
		switch {
		case install.Installed(it):
			r.Logf("✓ %s already installed", it.Name)
		case !install.Available(it):
			r.Logf("- skipping %s: %s not found", it.Name, install.Tool(it))
			skipped = append(skipped, it.Name)
		default:
			r.Logf("→ installing %s", it.Name)
			if err := install.Install(r, it); err != nil {
				r.Logf("✗ %s: %v", it.Name, err)
				failed = append(failed, it.Name)
			}
		}
	}
	return failed, skipped
}

const ohMyZshInstaller = "https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh"

func OhMyZsh(r *install.Runner, home string) error {
	if _, err := os.Stat(filepath.Join(home, ".oh-my-zsh")); err == nil {
		r.Logf("✓ Oh My Zsh already installed")
		return nil
	}
	return r.Run("sh", "-c", "curl -fsSL "+ohMyZshInstaller+" | sh -s -- --unattended --keep-zshrc")
}

func LoginShell(r *install.Runner) error {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		return err
	}
	if filepath.Base(os.Getenv("SHELL")) == "zsh" {
		r.Logf("✓ zsh is already the login shell")
		return nil
	}
	return r.Run("chsh", "-s", zsh)
}

// Link is one entry of dotfiles/links.toml.
type Link struct {
	Src string `toml:"src"` // relative to the repo root
	Dst string `toml:"dst"` // may start with ~/
}

// Dotfiles symlinks every entry of dotfiles/links.toml into home, moving any
// existing file out of the way into ~/.init-backup/<timestamp>/.
func Dotfiles(r *install.Runner, repo, home string, now time.Time) error {
	var cfg struct {
		Links []Link `toml:"link"`
	}
	if _, err := toml.DecodeFile(filepath.Join(repo, "dotfiles", "links.toml"), &cfg); err != nil {
		return err
	}
	backup := filepath.Join(home, ".init-backup", now.Format("20060102-150405"))
	for _, l := range cfg.Links {
		if err := link(r, filepath.Join(repo, l.Src), expandHome(l.Dst, home), home, backup); err != nil {
			return err
		}
	}
	return nil
}

func link(r *install.Runner, src, dst, home, backup string) error {
	if target, err := os.Readlink(dst); err == nil && target == src {
		r.Logf("✓ %s already linked", dst)
		return nil
	}
	if _, err := os.Lstat(dst); err == nil {
		rel, err := filepath.Rel(home, dst)
		if err != nil {
			return err
		}
		saved := filepath.Join(backup, rel)
		r.Logf("→ backing up %s to %s", dst, saved)
		if !r.DryRun {
			if err := os.MkdirAll(filepath.Dir(saved), 0o755); err != nil {
				return err
			}
			if err := os.Rename(dst, saved); err != nil {
				return err
			}
		}
	}
	r.Logf("→ linking %s -> %s", dst, src)
	if r.DryRun {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Symlink(src, dst)
}

func expandHome(path, home string) string {
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, rest)
	}
	return path
}
