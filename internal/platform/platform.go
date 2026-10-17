// Package platform detects which supported environment the installer runs on.
package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type OS string

const (
	MacOS  OS = "macos"
	Arch   OS = "arch"
	Ubuntu OS = "ubuntu"
)

type Platform struct {
	OS  OS
	WSL bool
}

func (p Platform) String() string {
	name := map[OS]string{MacOS: "macOS", Arch: "Arch Linux", Ubuntu: "Ubuntu"}[p.OS]
	if p.WSL {
		name += " (WSL)"
	}
	return name
}

// Detect returns the platform of the running machine.
func Detect() (Platform, error) {
	osRelease, _ := os.ReadFile("/etc/os-release")
	procVersion, _ := os.ReadFile("/proc/version")
	return detect(runtime.GOOS, string(osRelease), string(procVersion))
}

func detect(goos, osRelease, procVersion string) (Platform, error) {
	switch goos {
	case "darwin":
		return Platform{OS: MacOS}, nil
	case "linux":
		wsl := strings.Contains(strings.ToLower(procVersion), "microsoft")
		ids := osReleaseIDs(osRelease)
		for _, id := range ids {
			switch id {
			case "arch":
				return Platform{OS: Arch, WSL: wsl}, nil
			case "ubuntu":
				return Platform{OS: Ubuntu, WSL: wsl}, nil
			}
		}
		return Platform{}, fmt.Errorf("unsupported Linux distribution %v", ids)
	}
	return Platform{}, fmt.Errorf("unsupported OS %q", goos)
}

// osReleaseIDs returns ID followed by the entries of ID_LIKE from an os-release file.
func osReleaseIDs(osRelease string) []string {
	var id, like []string
	for _, line := range strings.Split(osRelease, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"'`)
		switch key {
		case "ID":
			id = []string{value}
		case "ID_LIKE":
			like = strings.Fields(value)
		}
	}
	return append(id, like...)
}
