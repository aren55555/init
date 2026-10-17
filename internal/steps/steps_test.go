package steps

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aren55555/init/internal/install"
)

func TestDotfiles(t *testing.T) {
	repo, home := t.TempDir(), t.TempDir()
	write(t, filepath.Join(repo, "dotfiles", "links.toml"), "[[link]]\nsrc = \"dotfiles/zsh/.zshrc\"\ndst = \"~/.zshrc\"\n")
	write(t, filepath.Join(repo, "dotfiles", "zsh", ".zshrc"), "# repo\n")
	write(t, filepath.Join(home, ".zshrc"), "# old\n")

	r := &install.Runner{Out: io.Discard}
	now := time.Date(2026, 10, 17, 15, 30, 0, 0, time.UTC)
	for range 2 { // the second run must be a no-op
		if err := Dotfiles(r, repo, home, now); err != nil {
			t.Fatal(err)
		}
	}

	target, err := os.Readlink(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(repo, "dotfiles", "zsh", ".zshrc"); target != want {
		t.Errorf("~/.zshrc -> %s, want %s", target, want)
	}
	saved, err := os.ReadFile(filepath.Join(home, ".init-backup", "20261017-153000", ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	if string(saved) != "# old\n" {
		t.Errorf("backup = %q", saved)
	}
}

func TestDotfilesDryRun(t *testing.T) {
	repo, home := t.TempDir(), t.TempDir()
	write(t, filepath.Join(repo, "dotfiles", "links.toml"), "[[link]]\nsrc = \"a\"\ndst = \"~/.a\"\n")
	write(t, filepath.Join(home, ".a"), "keep\n")

	if err := Dotfiles(&install.Runner{DryRun: true, Out: io.Discard}, repo, home, time.Now()); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(home, ".a")); string(got) != "keep\n" {
		t.Errorf("dry run changed ~/.a: %q", got)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
