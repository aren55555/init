package steps

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestAlfredHotkeyNotOpenedYet(t *testing.T) {
	var out strings.Builder
	if err := setAlfredHotkey(&install.Runner{Out: &out}, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hasn't been opened yet") {
		t.Errorf("output = %q", out.String())
	}
}

func TestAlfredHotkeyAlreadySet(t *testing.T) {
	if _, err := exec.LookPath("plutil"); err != nil {
		t.Skip("needs plutil (macOS)")
	}
	home := t.TempDir()
	prefsDir := filepath.Join(home, "sync", "Alfred.alfredpreferences")
	write(t, filepath.Join(home, "Library", "Application Support", "Alfred", "prefs.json"),
		`{"current": "`+strings.ReplaceAll(prefsDir, "/", `\/`)+`", "localhash": "abc123"}`)
	path, ok := alfredHotkeyPath(home)
	if want := filepath.Join(prefsDir, "preferences", "local", "abc123", "hotkey", "prefs.plist"); !ok || path != want {
		t.Fatalf("alfredHotkeyPath = %q, %v; want %q", path, ok, want)
	}
	if err := writeAlfredHotkey(path); err != nil {
		t.Fatal(err)
	}

	// With ⌘Space already set it must not quit, edit or reopen Alfred.
	var out strings.Builder
	if err := setAlfredHotkey(&install.Runner{Out: &out}, home); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "✓ Alfred already uses ⌘Space\n" {
		t.Errorf("output = %q", got)
	}
}

func TestMacSettingsDryRun(t *testing.T) {
	if _, err := exec.LookPath("defaults"); err != nil {
		t.Skip("needs defaults (macOS)")
	}
	// A key that can't exist yet, so the step must want to write it.
	d := Default{"test", "NSGlobalDomain", "com.aren55555.init.test-" + t.Name(), 15}
	var out strings.Builder
	if err := MacSettings(&install.Runner{DryRun: true, Out: &out}, []Default{d}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "$ defaults write NSGlobalDomain "+d.Key+" -int 15\n") {
		t.Errorf("output = %q", out.String())
	}
}
