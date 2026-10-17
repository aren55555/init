package platform

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name, goos, osRelease, procVersion string
		want                               Platform
	}{
		{"macos", "darwin", "", "", Platform{OS: MacOS}},
		{"omarchy", "linux", "NAME=\"Arch Linux\"\nID=arch\n", "Linux version 6.9.1-arch1-1", Platform{OS: Arch}},
		{"ubuntu", "linux", "NAME=\"Ubuntu\"\nID=ubuntu\nID_LIKE=debian\n", "Linux version 6.8.0-45-generic", Platform{OS: Ubuntu}},
		{"wsl", "linux", "ID=ubuntu\nID_LIKE=debian\n", "Linux version 5.15.153.1-microsoft-standard-WSL2", Platform{OS: Ubuntu, WSL: true}},
		{"ubuntu derivative", "linux", "ID=pop\nID_LIKE=\"ubuntu debian\"\n", "", Platform{OS: Ubuntu}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detect(tt.goos, tt.osRelease, tt.procVersion)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDetectUnsupported(t *testing.T) {
	if _, err := detect("linux", "ID=fedora\n", ""); err == nil {
		t.Error("expected error for fedora")
	}
	if _, err := detect("windows", "", ""); err == nil {
		t.Error("expected error for windows")
	}
}
