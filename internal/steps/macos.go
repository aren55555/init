package steps

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aren55555/init/internal/install"
)

// Default is one macOS `defaults` setting. Value is a bool or an int.
type Default struct {
	Name, Domain, Key string
	Value             any
}

// MacDefaults are the macOS settings the installer applies.
var MacDefaults = []Default{
	{"natural scrolling off", "NSGlobalDomain", "com.apple.swipescrolldirection", false},
	// The fastest rate and shortest delay System Settings offers.
	{"fastest key repeat", "NSGlobalDomain", "KeyRepeat", 2},
	{"shortest delay until key repeat", "NSGlobalDomain", "InitialKeyRepeat", 15},
	{"shake to locate pointer off", "NSGlobalDomain", "CGDisableCursorLocationMagnification", true},
}

// MacSettings (macOS only) applies MacDefaults, writing only those that differ.
func MacSettings(r *install.Runner, settings []Default) error {
	changed := false
	for _, d := range settings {
		typ, want, read := defaultArgs(d.Value)
		out, _ := exec.Command("defaults", "read", d.Domain, d.Key).Output()
		if strings.TrimSpace(string(out)) == read {
			r.Logf("✓ %s", d.Name)
			continue
		}
		r.Logf("→ %s", d.Name)
		if err := r.Run("defaults", "write", d.Domain, d.Key, typ, want); err != nil {
			return err
		}
		changed = true
	}
	if !changed {
		return nil
	}
	// Some of these only take effect at the next login without this.
	if _, err := os.Stat(activateSettings); err != nil {
		r.Logf("→ log out and back in for every setting to take effect")
		return nil
	}
	return r.Run(activateSettings, "-u")
}

// defaultArgs returns the `defaults write` type flag and value for v, and
// what `defaults read` prints once it's set.
func defaultArgs(v any) (typ, write, read string) {
	switch v := v.(type) {
	case bool:
		if v {
			return "-bool", "true", "1"
		}
		return "-bool", "false", "0"
	case int:
		s := strconv.Itoa(v)
		return "-int", s, s
	}
	panic(fmt.Sprintf("unsupported defaults value %T", v))
}
