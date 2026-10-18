package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aren55555/init/internal/install"
	"github.com/aren55555/init/internal/manifest"
)

const (
	alfredBundle = "com.runningwithcrayons.Alfred"
	// ⌘Space: virtual key code 49 (space) with the ⌘ modifier mask.
	spaceKey, cmdMod = 49, 1048576
	// "Show Spotlight search" in com.apple.symbolichotkeys.
	spotlightHotkey  = "64"
	activateSettings = "/System/Library/PrivateFrameworks/SystemAdministration.framework/Resources/activateSettings"
)

// AlfredHotkey (macOS only) moves ⌘Space from Spotlight to Alfred. Alfred
// keeps its hotkey under a per-machine hash it creates on first launch, so
// before then this only frees ⌘Space and says how to finish.
func AlfredHotkey(r *install.Runner, home string) error {
	if !install.Installed(manifest.Item{Method: manifest.Cask, Pkg: "alfred"}) {
		r.Logf("- skipping ⌘Space for Alfred: Alfred isn't installed")
		return nil
	}
	if err := disableSpotlightHotkey(r, home); err != nil {
		return err
	}
	return setAlfredHotkey(r, home)
}

func disableSpotlightHotkey(r *install.Runner, home string) error {
	plist := filepath.Join(home, "Library", "Preferences", "com.apple.symbolichotkeys.plist")
	out, _ := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print :AppleSymbolicHotKeys:"+spotlightHotkey+":enabled", plist).Output()
	if strings.TrimSpace(string(out)) == "false" {
		r.Logf("✓ Spotlight's ⌘Space shortcut already off")
		return nil
	}
	entry := fmt.Sprintf("<dict><key>enabled</key><false/><key>value</key><dict><key>parameters</key>"+
		"<array><integer>32</integer><integer>%d</integer><integer>%d</integer></array>"+
		"<key>type</key><string>standard</string></dict></dict>", spaceKey, cmdMod)
	if err := r.Run("defaults", "write", "com.apple.symbolichotkeys", "AppleSymbolicHotKeys", "-dict-add", spotlightHotkey, entry); err != nil {
		return err
	}
	// Applies the change now rather than at the next login.
	if _, err := os.Stat(activateSettings); err != nil {
		r.Logf("→ Spotlight's ⌘Space shortcut turns off after you log out and back in")
		return nil
	}
	return r.Run(activateSettings, "-u")
}

func setAlfredHotkey(r *install.Runner, home string) error {
	path, ok := alfredHotkeyPath(home)
	if !ok {
		r.Logf("→ Alfred hasn't been opened yet: open it and pick ⌘Space as its hotkey, or re-run init afterwards")
		return nil
	}
	if plistInt(path, "default.key") == spaceKey && plistInt(path, "default.mod") == cmdMod {
		r.Logf("✓ Alfred already uses ⌘Space")
		return nil
	}
	// Alfred saves its preferences when it quits, so quit it before editing them.
	if err := r.Run("osascript", "-e", fmt.Sprintf("if application id %q is running then quit application id %q", alfredBundle, alfredBundle)); err != nil {
		return err
	}
	r.Logf("→ setting Alfred's hotkey to ⌘Space in %s", path)
	if !r.DryRun {
		if err := writeAlfredHotkey(path); err != nil {
			return err
		}
	}
	return r.Run("open", "-b", alfredBundle)
}

// alfredHotkeyPath finds Alfred's hotkey preferences through its prefs.json,
// which names the preferences folder (possibly a synced one) and this machine's hash.
func alfredHotkeyPath(home string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(home, "Library", "Application Support", "Alfred", "prefs.json"))
	if err != nil {
		return "", false
	}
	var prefs struct {
		Current   string `json:"current"`
		LocalHash string `json:"localhash"`
	}
	if json.Unmarshal(data, &prefs) != nil || prefs.Current == "" || prefs.LocalHash == "" {
		return "", false
	}
	return filepath.Join(prefs.Current, "preferences", "local", prefs.LocalHash, "hotkey", "prefs.plist"), true
}

func writeAlfredHotkey(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, fmt.Appendf(nil, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>default</key>
	<dict>
		<key>key</key>
		<integer>%d</integer>
		<key>mod</key>
		<integer>%d</integer>
		<key>string</key>
		<string> </string>
	</dict>
</dict>
</plist>
`, spaceKey, cmdMod), 0o644)
}

// plistInt reads an integer at a dotted key path, or -1 if it can't.
func plistInt(path, key string) int {
	out, err := exec.Command("plutil", "-extract", key, "raw", path).Output()
	if err != nil {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return -1
	}
	return n
}
