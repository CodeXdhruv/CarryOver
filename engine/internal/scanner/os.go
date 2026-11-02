package scanner

import (
	"bufio"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// OSInfo captures detected operating system identity and package manager
type OSInfo struct {
	ID             string `json:"id"`
	IDLike         string `json:"id_like,omitempty"`
	VersionID      string `json:"version_id,omitempty"`
	PrettyName     string `json:"pretty_name,omitempty"`
	PackageManager string `json:"package_manager"`
}

// SystemEnvironment captures runtime system context
type SystemEnvironment struct {
	Hostname string
	Username string
	Home     string
	Shell    string
	Desktop  string
	Arch     string
}

// ParseOSRelease parses key-value pairs from /etc/os-release reader
func ParseOSRelease(r io.Reader) (*OSInfo, error) {
	scanner := bufio.NewScanner(r)
	info := &OSInfo{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "ID":
			info.ID = strings.ToLower(val)
		case "ID_LIKE":
			info.IDLike = strings.ToLower(val)
		case "VERSION_ID":
			info.VersionID = val
		case "PRETTY_NAME":
			info.PrettyName = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	info.PackageManager = DetectPackageManager(info.ID, info.IDLike)
	return info, nil
}

// DetectPackageManager maps OS ID and ID_LIKE to the primary package manager
func DetectPackageManager(id, idLike string) string {
	id = strings.ToLower(id)
	idLike = strings.ToLower(idLike)
	combined := id + " " + idLike

	switch {
	case strings.Contains(combined, "ubuntu") ||
		strings.Contains(combined, "debian") ||
		strings.Contains(combined, "linuxmint") ||
		strings.Contains(combined, "pop"):
		return "apt"
	case strings.Contains(combined, "fedora") ||
		strings.Contains(combined, "rhel") ||
		strings.Contains(combined, "centos") ||
		strings.Contains(combined, "rocky"):
		return "dnf"
	case strings.Contains(combined, "arch") ||
		strings.Contains(combined, "manjaro") ||
		strings.Contains(combined, "endeavouros"):
		return "pacman"
	case strings.Contains(combined, "opensuse") ||
		strings.Contains(combined, "suse") ||
		strings.Contains(combined, "sles"):
		return "zypper"
	default:
		return "unknown"
	}
}

// DetectOS reads /etc/os-release with optional custom root path
func DetectOS(rootDir string) (*OSInfo, error) {
	if rootDir == "" {
		rootDir = "/"
	}
	osReleasePath := filepath.Join(rootDir, "etc", "os-release")
	f, err := os.Open(osReleasePath)
	if err != nil {
		// Try usr/lib/os-release fallback
		osReleasePath = filepath.Join(rootDir, "usr", "lib", "os-release")
		f, err = os.Open(osReleasePath)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()
	return ParseOSRelease(f)
}

// DetectSystemEnvironment captures current host, user, shell, desktop and architecture
func DetectSystemEnvironment() SystemEnvironment {
	env := SystemEnvironment{
		Arch: runtime.GOARCH,
	}

	if h, err := os.Hostname(); err == nil {
		env.Hostname = h
	}

	if u, err := user.Current(); err == nil {
		env.Username = u.Username
		env.Home = u.HomeDir
	} else {
		env.Username = os.Getenv("USER")
		env.Home = os.Getenv("HOME")
	}

	env.Shell = os.Getenv("SHELL")
	if env.Shell == "" {
		env.Shell = "/bin/bash"
	}

	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = os.Getenv("DESKTOP_SESSION")
	}
	env.Desktop = desktop

	return env
}
