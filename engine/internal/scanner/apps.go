package scanner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/carryover/carryover/engine/internal/manifest"
)

// CommandRunner executes a command and returns its stdout
type CommandRunner func(name string, args ...string) ([]byte, error)

// DefaultCommandRunner uses os/exec
func DefaultCommandRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// ParseAptManual parses output of apt-mark showmanual
func ParseAptManual(r io.Reader) []manifest.PkgItem {
	var items []manifest.PkgItem
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pkg := strings.TrimSpace(scanner.Text())
		if pkg == "" || strings.HasPrefix(pkg, "#") {
			continue
		}
		items = append(items, manifest.PkgItem{Name: pkg})
	}
	return items
}

// ParseDnfUserinstalled parses output of dnf repoquery --userinstalled
// e.g. git-0:2.44.0-1.fc40.x86_64 -> name: git, version: 2.44.0-1.fc40
func ParseDnfUserinstalled(r io.Reader) []manifest.PkgItem {
	var items []manifest.PkgItem
	scanner := bufio.NewScanner(r)
	re := regexp.MustCompile(`^(.+?)(?:-\d+:|-)(.+?)\.[^.]+$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			items = append(items, manifest.PkgItem{
				Name:    matches[1],
				Version: matches[2],
			})
		} else {
			items = append(items, manifest.PkgItem{Name: line})
		}
	}
	return items
}

// ParsePacmanExplicit parses output of pacman -Qqe or -Qqm
func ParsePacmanExplicit(r io.Reader, repo string) []manifest.PacmanItem {
	var items []manifest.PacmanItem
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Some formats have name and version: pkgname 1.2.3-1
		parts := strings.Fields(line)
		name := parts[0]
		version := ""
		if len(parts) > 1 {
			version = parts[1]
		}
		items = append(items, manifest.PacmanItem{
			Name:    name,
			Version: version,
			Repo:    repo,
		})
	}
	return items
}

// ParseZypperInstalled parses user-installed packages from zypper
func ParseZypperInstalled(r io.Reader) []manifest.PkgItem {
	var items []manifest.PkgItem
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: i+ | package-name | ...
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			pkg := strings.TrimSpace(parts[1])
			if pkg != "" && pkg != "Name" {
				items = append(items, manifest.PkgItem{Name: pkg})
			}
		} else {
			items = append(items, manifest.PkgItem{Name: line})
		}
	}
	return items
}

// ParseFlatpakList parses output of flatpak list --app --columns=application,origin
func ParseFlatpakList(r io.Reader) []manifest.FlatpakItem {
	var items []manifest.FlatpakItem
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			items = append(items, manifest.FlatpakItem{
				ID:     fields[0],
				Origin: fields[1],
			})
		} else if len(fields) == 1 {
			items = append(items, manifest.FlatpakItem{
				ID: fields[0],
			})
		}
	}
	return items
}

// ParseSnapList parses output of snap list, skipping core/base/snapd
func ParseSnapList(r io.Reader) []manifest.SnapItem {
	var items []manifest.SnapItem
	scanner := bufio.NewScanner(r)
	headerSkipped := false

	skipNames := map[string]bool{
		"bare":               true,
		"core":               true,
		"core18":             true,
		"core20":             true,
		"core22":             true,
		"core24":             true,
		"snapd":              true,
		"canonical-livepatch": true,
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !headerSkipped {
			if strings.HasPrefix(line, "Name") {
				headerSkipped = true
				continue
			}
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		version := fields[1]
		if skipNames[name] {
			continue
		}
		classic := false
		for _, f := range fields {
			if f == "classic" {
				classic = true
				break
			}
		}
		items = append(items, manifest.SnapItem{
			Name:    name,
			Version: version,
			Classic: classic,
		})
	}
	return items
}

// ParsePipList parses output of pip list --user --format=json
func ParsePipList(r io.Reader) []string {
	type pipPkg struct {
		Name string `json:"name"`
	}
	var pkgs []pipPkg
	if err := json.NewDecoder(r).Decode(&pkgs); err != nil {
		return nil
	}
	var res []string
	for _, p := range pkgs {
		if p.Name != "" {
			res = append(res, p.Name)
		}
	}
	return res
}

// ParseNpmGlobal parses output of npm ls -g --depth=0 --json
func ParseNpmGlobal(r io.Reader) []string {
	type npmOutput struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}
	var out npmOutput
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil
	}
	var res []string
	for k := range out.Dependencies {
		res = append(res, k)
	}
	return res
}

// ParseCargoList parses output of cargo install --list
func ParseCargoList(r io.Reader) []string {
	var items []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		// Format: ripgrep v14.1.0:
		fields := strings.Fields(line)
		if len(fields) > 0 {
			name := strings.TrimSuffix(fields[0], ":")
			if name != "" {
				items = append(items, name)
			}
		}
	}
	return items
}

// ParseVSCodeExtensions parses lines from code --list-extensions
func ParseVSCodeExtensions(r io.Reader) []string {
	var items []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			items = append(items, line)
		}
	}
	return items
}

// ScanApps performs package and app discovery based on the target OS package manager
func ScanApps(runner CommandRunner, pm string) manifest.AppSet {
	if runner == nil {
		runner = DefaultCommandRunner
	}
	var appSet manifest.AppSet

	switch pm {
	case "apt":
		if out, err := runner("apt-mark", "showmanual"); err == nil {
			appSet.Apt = ParseAptManual(bytes.NewReader(out))
		}
	case "dnf":
		if out, err := runner("dnf", "repoquery", "--userinstalled"); err == nil {
			appSet.Dnf = ParseDnfUserinstalled(bytes.NewReader(out))
		}
	case "pacman":
		if out, err := runner("pacman", "-Qqe"); err == nil {
			appSet.Pacman = ParsePacmanExplicit(bytes.NewReader(out), "")
		}
		if aurOut, err := runner("pacman", "-Qqm"); err == nil {
			aurPkgs := ParsePacmanExplicit(bytes.NewReader(aurOut), "AUR")
			appSet.Pacman = append(appSet.Pacman, aurPkgs...)
		}
	case "zypper":
		if out, err := runner("zypper", "se", "-i"); err == nil {
			appSet.Zypper = ParseZypperInstalled(bytes.NewReader(out))
		}
	}

	// Flatpak
	if out, err := runner("flatpak", "list", "--app", "--columns=application,origin"); err == nil {
		appSet.Flatpak = ParseFlatpakList(bytes.NewReader(out))
	}

	// Snap
	if out, err := runner("snap", "list"); err == nil {
		appSet.Snap = ParseSnapList(bytes.NewReader(out))
	}

	// Pip user
	if out, err := runner("pip", "list", "--user", "--format=json"); err == nil {
		appSet.PipUser = ParsePipList(bytes.NewReader(out))
	}

	// NPM global
	if out, err := runner("npm", "ls", "-g", "--depth=0", "--json"); err == nil {
		appSet.NpmGlobal = ParseNpmGlobal(bytes.NewReader(out))
	}

	// Cargo
	if out, err := runner("cargo", "install", "--list"); err == nil {
		appSet.Cargo = ParseCargoList(bytes.NewReader(out))
	}

	// VS Code
	if out, err := runner("code", "--list-extensions"); err == nil {
		appSet.VSCodeExtensions = ParseVSCodeExtensions(bytes.NewReader(out))
	}

	return appSet
}
