// Package install runs commands and installs manifest items with the right package manager.
package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aren55555/init/internal/manifest"
)

// Runner runs commands with the terminal attached, or only prints them when DryRun is set.
type Runner struct {
	DryRun bool
	Out    io.Writer
}

func (r *Runner) Logf(format string, args ...any) {
	fmt.Fprintf(r.Out, format+"\n", args...)
}

func (r *Runner) Run(name string, args ...string) error {
	r.Logf("$ %s", strings.Join(append([]string{name}, args...), " "))
	if r.DryRun {
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// Sudo runs a command as root, prefixing sudo unless already root.
func (r *Runner) Sudo(name string, args ...string) error {
	if os.Geteuid() == 0 {
		return r.Run(name, args...)
	}
	return r.Run("sudo", append([]string{name}, args...)...)
}

// succeeds reports whether a read-only check command exits 0.
func succeeds(name string, args ...string) bool {
	return exec.Command(name, args...).Run() == nil
}

// OnPath reports whether a command is on PATH or in ~/.local/bin, where
// script installers like mise's put binaries before PATH knows about them.
func OnPath(name string) bool {
	if _, err := exec.LookPath(name); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".local", "bin", name))
	return err == nil
}

// Tool returns the command that installs it.
func Tool(it manifest.Item) string {
	return map[manifest.Method]string{
		manifest.Brew:   "brew",
		manifest.Cask:   "brew",
		manifest.Pacman: "pacman",
		manifest.AUR:    "yay",
		manifest.Apt:    "apt-get",
		manifest.Script: "sh",
	}[it.Method]
}

// Available reports whether the package manager for it is present.
func Available(it manifest.Item) bool {
	return OnPath(Tool(it))
}

func Installed(it manifest.Item) bool {
	switch it.Method {
	case manifest.Brew:
		return succeeds("brew", "list", "--formula", it.Pkg)
	case manifest.Cask:
		return succeeds("brew", "list", "--cask", it.Pkg)
	case manifest.Pacman, manifest.AUR:
		return succeeds("pacman", "-Q", it.Pkg)
	case manifest.Apt:
		return succeeds("dpkg", "-s", it.Pkg)
	case manifest.Script:
		return OnPath(it.Check)
	}
	return false
}

func Install(r *Runner, it manifest.Item) error {
	switch it.Method {
	// HOMEBREW_NO_ASK: Homebrew 7 otherwise stops at "Do you want to proceed with
	// the installation? [y/n]" whenever the plan pulls in dependencies.
	case manifest.Brew:
		return r.Run("env", "HOMEBREW_NO_ASK=1", "brew", "install", it.Pkg)
	case manifest.Cask:
		return r.Run("env", "HOMEBREW_NO_ASK=1", "brew", "install", "--cask", it.Pkg)
	case manifest.Pacman:
		return r.Sudo("pacman", "-S", "--needed", "--noconfirm", it.Pkg)
	case manifest.AUR:
		return r.Run("yay", "-S", "--needed", "--noconfirm", it.Pkg)
	case manifest.Apt:
		return r.Sudo("apt-get", "install", "-y", it.Pkg)
	case manifest.Script:
		return r.Run("sh", "-c", it.Pkg)
	}
	return fmt.Errorf("unknown install method %q", it.Method)
}
