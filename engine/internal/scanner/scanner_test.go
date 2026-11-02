package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/carryover/carryover/engine/internal/events"
)

func TestDetectOSFromFixtures(t *testing.T) {
	fixtures := []struct {
		file       string
		expectedID string
		expectedPM string
	}{
		{"ubuntu", "ubuntu", "apt"},
		{"fedora", "fedora", "dnf"},
		{"arch", "arch", "pacman"},
		{"opensuse", "opensuse-tumbleweed", "zypper"},
	}

	for _, tc := range fixtures {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "tests", "fixtures", "os-release", tc.file)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to open fixture %s: %v", path, err)
			}
			defer f.Close()

			info, err := ParseOSRelease(f)
			if err != nil {
				t.Fatalf("ParseOSRelease error: %v", err)
			}
			if info.ID != tc.expectedID {
				t.Errorf("expected ID %s, got %s", tc.expectedID, info.ID)
			}
			if info.PackageManager != tc.expectedPM {
				t.Errorf("expected PM %s, got %s", tc.expectedPM, info.PackageManager)
			}
		})
	}
}

func TestAppParsers(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "..", "tests", "fixtures", "apps")

	// Apt
	aptData, err := os.ReadFile(filepath.Join(fixturesDir, "apt-manual.txt"))
	if err != nil {
		t.Fatalf("failed to read apt fixture: %v", err)
	}
	aptItems := ParseAptManual(bytes.NewReader(aptData))
	if len(aptItems) < 5 {
		t.Errorf("expected >= 5 apt items, got %d", len(aptItems))
	}

	// Dnf
	dnfData, err := os.ReadFile(filepath.Join(fixturesDir, "dnf-userinstalled.txt"))
	if err != nil {
		t.Fatalf("failed to read dnf fixture: %v", err)
	}
	dnfItems := ParseDnfUserinstalled(bytes.NewReader(dnfData))
	if len(dnfItems) != 5 {
		t.Errorf("expected 5 dnf items, got %d", len(dnfItems))
	}
	if dnfItems[0].Name != "git" || dnfItems[0].Version != "2.44.0-1.fc40" {
		t.Errorf("unexpected first dnf item: %+v", dnfItems[0])
	}

	// Pacman
	pacData, err := os.ReadFile(filepath.Join(fixturesDir, "pacman-explicit.txt"))
	if err != nil {
		t.Fatalf("failed to read pacman fixture: %v", err)
	}
	pacItems := ParsePacmanExplicit(bytes.NewReader(pacData), "")
	if len(pacItems) < 5 {
		t.Errorf("expected >= 5 pacman items, got %d", len(pacItems))
	}

	// Flatpak
	flatData, err := os.ReadFile(filepath.Join(fixturesDir, "flatpak-list.txt"))
	if err != nil {
		t.Fatalf("failed to read flatpak fixture: %v", err)
	}
	flatItems := ParseFlatpakList(bytes.NewReader(flatData))
	if len(flatItems) != 3 {
		t.Errorf("expected 3 flatpak items, got %d", len(flatItems))
	}
	if flatItems[0].ID != "org.telegram.desktop" || flatItems[0].Origin != "flathub" {
		t.Errorf("unexpected flatpak item: %+v", flatItems[0])
	}

	// Snap
	snapData, err := os.ReadFile(filepath.Join(fixturesDir, "snap-list.txt"))
	if err != nil {
		t.Fatalf("failed to read snap fixture: %v", err)
	}
	snapItems := ParseSnapList(bytes.NewReader(snapData))
	// Should keep code and spotify, skip bare, core22, snapd
	if len(snapItems) != 2 {
		t.Fatalf("expected 2 snap items after filtering, got %d", len(snapItems))
	}
	if snapItems[0].Name != "code" || !snapItems[0].Classic {
		t.Errorf("expected code with classic=true, got: %+v", snapItems[0])
	}
	if snapItems[1].Name != "spotify" {
		t.Errorf("expected spotify, got: %+v", snapItems[1])
	}
}

func TestShellRCParser(t *testing.T) {
	path := filepath.Join("..", "..", "..", "tests", "fixtures", "shell", "sample.bashrc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read bashrc fixture: %v", err)
	}

	env, aliases, paths := ParseShellRC(bytes.NewReader(data))
	if env["EDITOR"] != "nvim" {
		t.Errorf("expected EDITOR=nvim, got %s", env["EDITOR"])
	}
	if aliases["ll"] != "ls -alF" {
		t.Errorf("expected ll alias, got %s", aliases["ll"])
	}
	if len(paths) == 0 {
		t.Errorf("expected at least 1 path addition")
	}
}

func TestProjectDetectionRecipes(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Node project
	nodeDir := filepath.Join(tmpDir, "node-app")
	_ = os.MkdirAll(nodeDir, 0755)
	_ = os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(nodeDir, "package-lock.json"), []byte("{}"), 0644)

	recipe := DetectProjectInDir(nodeDir)
	if recipe == nil || recipe.Type != "node" || recipe.Rebuild[0] != "npm ci" {
		t.Fatalf("unexpected node recipe: %+v", recipe)
	}

	// 2. Rust project
	rustDir := filepath.Join(tmpDir, "rust-app")
	_ = os.MkdirAll(rustDir, 0755)
	_ = os.WriteFile(filepath.Join(rustDir, "Cargo.toml"), []byte(""), 0644)

	recipeRust := DetectProjectInDir(rustDir)
	if recipeRust == nil || recipeRust.Type != "rust" || recipeRust.Rebuild[0] != "cargo fetch" {
		t.Fatalf("unexpected rust recipe: %+v", recipeRust)
	}

	// 3. ScanProjects walk
	projects, perProjectExcludes := ScanProjects(tmpDir, []string{"."}, 2)
	if len(projects) != 2 {
		t.Errorf("expected 2 detected projects, got %d", len(projects))
	}
	if len(perProjectExcludes) != 2 {
		t.Errorf("expected 2 per-project exclude mappings, got %d", len(perProjectExcludes))
	}
}

func TestScanSystemIntegration(t *testing.T) {
	tmpHome := t.TempDir()

	// Seed folders
	docsDir := filepath.Join(tmpHome, "Documents")
	_ = os.MkdirAll(docsDir, 0755)
	_ = os.WriteFile(filepath.Join(docsDir, "resume.pdf"), []byte("pdf content"), 0644)

	// Seed dotfiles
	_ = os.WriteFile(filepath.Join(tmpHome, ".bashrc"), []byte("export EDITOR=nvim\nalias ll='ls -l'\n"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpHome, ".ssh"), 0700)
	_ = os.WriteFile(filepath.Join(tmpHome, ".ssh", "config"), []byte("Host *\n"), 0600)

	// Seed project
	projDir := filepath.Join(tmpHome, "Projects", "my-app")
	_ = os.MkdirAll(projDir, 0755)
	_ = os.WriteFile(filepath.Join(projDir, "package.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(projDir, "yarn.lock"), []byte(""), 0644)

	// Mock command runner
	mockRunner := func(name string, args ...string) ([]byte, error) {
		if name == "apt-mark" {
			return []byte("git\ncurl\nbuild-essential\n"), nil
		}
		if name == "flatpak" {
			return []byte("org.videolan.VLC\tflathub\n"), nil
		}
		return nil, os.ErrNotExist
	}

	var buf bytes.Buffer
	emitter := events.NewEmitter(&buf, true)

	opts := ScanOptions{
		HomeDir:       tmpHome,
		CommandRunner: mockRunner,
		Emitter:       emitter,
	}

	m, err := ScanSystem(opts)
	if err != nil {
		t.Fatalf("ScanSystem error: %v", err)
	}

	if len(m.Folders) == 0 {
		t.Errorf("expected folders in manifest")
	}
	if len(m.Dotfiles) == 0 {
		t.Errorf("expected dotfiles in manifest")
	}
	if len(m.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(m.Projects))
	}
	if len(m.Apps.Apt) != 3 {
		t.Errorf("expected 3 apt apps, got %d", len(m.Apps.Apt))
	}
	if len(m.Apps.Flatpak) != 1 {
		t.Errorf("expected 1 flatpak app, got %d", len(m.Apps.Flatpak))
	}

	// Verify events were emitted
	if buf.Len() == 0 {
		t.Errorf("expected events emitted on stream")
	}

	// Verify manifest schema validation
	if err := m.Validate(); err != nil {
		t.Errorf("expected valid manifest schema: %v", err)
	}
}
