// Command vmtest runs bootstrap.sh in a throwaway Tart macOS VM, runs it a
// second time to check it's idempotent, then checks the machine's end state.
//
//	go run ./cmd/vmtest [-remote] [-keep]
//
// The first run downloads a vanilla macOS image (~24 GB) into a local VM named
// -base. That VM is never booted; each test boots a copy-on-write clone of it.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aren55555/init/internal/manifest"
	"github.com/aren55555/init/internal/platform"
)

const (
	// Default credentials of Cirrus Labs' vanilla images.
	guestUser     = "admin"
	guestPassword = "admin"

	// Where `tart run --dir=init:...` mounts the checkout inside a macOS guest.
	guestRepo = "/Volumes/My Shared Files/init"

	oneLiner = "curl -fsSL https://init.aren55555.com | sh -s -- --yes"
)

type check struct{ name, cmd string }

// checks run in the guest after bootstrapping; each passes if it exits 0.
// Every macOS item in packages.toml gets an "installed" check too.
var checks = []check{
	{"tools on PATH in a new shell", `zsh -lic 'for c in brew git gh mise zoxide go; do command -v $c >/dev/null || { echo "missing: $c"; exit 1; }; done'`},
	{"~/.zshrc links into the repo", `t=$(readlink ~/.zshrc) && case "$t" in */dotfiles/zsh/.zshrc) test -f "$t" ;; *) echo "~/.zshrc -> $t"; false ;; esac`},
	{"Oh My Zsh loads", `zsh -lic 'whence -w omz'`},
	{"prompt uses the agnoster theme", `zsh -lic '[[ $ZSH_THEME == agnoster ]] && whence -w build_prompt'`},
	{"zsh is the login shell", `dscl . -read "/Users/$USER" UserShell | grep -q zsh`},
}

// installedCheck verifies a manifest item in the guest the same way the installer does.
func installedCheck(it manifest.Item) check {
	cmd := fmt.Sprintf(`zsh -lic 'command -v %s'`, it.Check)
	switch it.Method {
	case manifest.Cask:
		cmd = fmt.Sprintf(`zsh -lic 'brew list --cask %s'`, it.Pkg)
	case manifest.Brew:
		cmd = fmt.Sprintf(`zsh -lic 'brew list --formula %s'`, it.Pkg)
	}
	return check{it.Name + " installed", cmd}
}

func main() {
	image := flag.String("image", "ghcr.io/cirruslabs/macos-tahoe-vanilla:latest", "OCI image the base VM is cloned from")
	base := flag.String("base", "mac-base", "local base VM, created from -image if missing")
	remote := flag.Bool("remote", false, "run the published one-liner instead of the local checkout")
	keep := flag.Bool("keep", false, "keep the test VM afterwards (open it with `tart run <name>`)")
	name := flag.String("name", "init-test-"+time.Now().Format("20060102-150405"), "name of the test VM")
	authorizedKey := flag.String("authorized-key", "", "SSH public key to authorize for the guest user")
	flag.Parse()
	if err := run(*image, *base, *name, *authorizedKey, *remote, *keep); err != nil {
		fmt.Fprintln(os.Stderr, "vmtest:", err)
		os.Exit(1)
	}
}

func run(image, base, name, authorizedKey string, remote, keep bool) error {
	repo, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(repo, "bootstrap.sh")); err != nil {
		return errors.New("run from the repo root")
	}
	m, err := manifest.Load(filepath.Join(repo, "packages.toml"))
	if err != nil {
		return err
	}
	allChecks := checks
	for _, it := range m.Resolve(platform.Platform{OS: platform.MacOS}) {
		allChecks = append(allChecks, installedCheck(it))
	}
	if authorizedKey != "" {
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey)); err != nil {
			return fmt.Errorf("-authorized-key: %w", err)
		}
	}
	logPath := filepath.Join(repo, ".logs", name+".log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()
	out := io.MultiWriter(os.Stdout, logFile)
	say := func(format string, args ...any) {
		fmt.Fprintf(out, "\n==> "+format+"\n", args...)
	}

	if exec.Command("tart", "get", base).Run() != nil {
		say("Creating %s from %s (one-time download, ~24 GB)", base, image)
		if err := tart(out, "clone", image, base); err != nil {
			return err
		}
	}

	say("Cloning %s into %s", base, name)
	if err := tart(out, "clone", base, name); err != nil {
		return err
	}
	if keep {
		defer say("Kept VM %s: open it with `tart run %s`, remove it with `tart delete %s`", name, name, name)
	} else {
		defer tart(out, "delete", name)
	}

	say("Booting %s", name)
	vm := exec.Command("tart", "run", "--no-graphics", "--dir=init:"+repo+":ro", name)
	vm.Stdout, vm.Stderr = logFile, logFile
	if err := vm.Start(); err != nil {
		return err
	}
	defer func() {
		tart(out, "stop", name)
		vm.Wait()
	}()

	ipOut, err := exec.Command("tart", "ip", "--wait", "300", name).Output()
	if err != nil {
		return fmt.Errorf("waiting for VM IP: %w", err)
	}
	g, err := dial(strings.TrimSpace(string(ipOut)))
	if err != nil {
		return err
	}
	defer g.Close()

	say("Enabling passwordless sudo, so installers don't stop for a password")
	if err := g.run(out, `echo `+guestPassword+` | sudo -S sh -c 'echo "`+guestUser+` ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/vmtest && chmod 440 /etc/sudoers.d/vmtest'`); err != nil {
		return err
	}
	if authorizedKey != "" {
		say("Authorizing SSH key")
		s, err := g.NewSession()
		if err != nil {
			return err
		}
		s.Stdin = strings.NewReader(strings.TrimSpace(authorizedKey) + "\n")
		s.Stdout, s.Stderr = out, out
		err = s.Run(`mkdir -p -m 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`)
		s.Close()
		if err != nil {
			return err
		}
	}

	bootstrap := fmt.Sprintf("cd %q && sh ./bootstrap.sh --yes", guestRepo)
	if remote {
		bootstrap = oneLiner
	}
	start := time.Now()
	say("First run: %s", bootstrap)
	firstErr := g.run(out, bootstrap)
	firstRun := time.Since(start).Round(time.Second)

	say("Second run (should change nothing)")
	var second bytes.Buffer
	secondErr := g.run(io.MultiWriter(out, &second), bootstrap)

	say("Checking the result")
	results := []result{{fmt.Sprintf("first run succeeds (took %s)", firstRun), firstErr == nil, fmt.Sprint(firstErr)}}
	// "→ " lines come from cmd/init; "==> Installing"/"==> Cloning" from bootstrap.sh or Homebrew.
	var changes []string
	for _, line := range strings.Split(second.String(), "\n") {
		if strings.HasPrefix(line, "→ ") || strings.HasPrefix(line, "==> Installing") || strings.HasPrefix(line, "==> Cloning") {
			changes = append(changes, line)
		}
	}
	secondDetail := strings.Join(changes, "; ")
	if secondErr != nil {
		secondDetail = secondErr.Error()
	}
	results = append(results, result{"second run succeeds and changes nothing", secondErr == nil && len(changes) == 0, secondDetail})
	for _, c := range allChecks {
		var buf bytes.Buffer
		err := g.run(io.MultiWriter(logFile, &buf), c.cmd)
		results = append(results, result{c.name, err == nil, strings.TrimSpace(buf.String())})
	}

	say("Results (log: %s)", logPath)
	failures := 0
	for _, r := range results {
		mark := "✓"
		if !r.ok {
			mark = "✗"
			failures++
		}
		fmt.Fprintf(out, "%s %s", mark, r.name)
		if !r.ok && r.detail != "" {
			fmt.Fprintf(out, ": %s", lastLine(r.detail))
		}
		fmt.Fprintln(out)
	}
	if failures > 0 {
		return fmt.Errorf("%d of %d checks failed", failures, len(results))
	}
	return nil
}

type result struct {
	name   string
	ok     bool
	detail string
}

type guest struct{ *ssh.Client }

// dial connects to the guest, retrying while sshd starts up.
func dial(ip string) (*guest, error) {
	cfg := &ssh.ClientConfig{
		User: guestUser,
		Auth: []ssh.AuthMethod{ssh.Password(guestPassword)},
		// Throwaway local VM with a fresh host key every run.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	deadline := time.Now().Add(3 * time.Minute)
	for {
		c, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), cfg)
		if err == nil {
			return &guest{c}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("ssh to %s: %w", ip, err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (g *guest) run(w io.Writer, cmd string) error {
	s, err := g.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	s.Stdout, s.Stderr = w, w
	return s.Run(cmd)
}

func tart(w io.Writer, args ...string) error {
	cmd := exec.Command("tart", args...)
	cmd.Stdout, cmd.Stderr = w, w
	return cmd.Run()
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}
