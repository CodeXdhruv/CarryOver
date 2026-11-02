package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestValidation(t *testing.T) {
	m := NewManifest()
	m.BackupMeta = BackupMeta{
		ID:               "2026-11-02T10-15-00Z",
		Created:          time.Now().UTC(),
		CarryoverVersion: "1.0.0",
		Source: SourceOS{
			Hostname:       "test-laptop",
			OSID:           "ubuntu",
			OSVersion:      "24.04",
			Arch:           "x86_64",
			PackageManager: "apt",
			Username:       "testuser",
			Home:           "/home/testuser",
			Shell:          "/bin/bash",
			Desktop:        "GNOME",
		},
	}
	m.Folders = []Folder{
		{Path: "Documents", SizeBytes: 1024},
		{Path: "Projects", SizeBytes: 2048},
	}
	m.Dotfiles = []Dotfile{
		{Path: ".bashrc"},
		{Path: ".ssh/config", Sensitive: true},
	}
	m.Apps.Apt = []PkgItem{
		{Name: "git"},
		{Name: "curl"},
	}
	m.Projects = []Project{
		{
			Path:           "Projects/web-app",
			Type:           "node",
			PackageManager: "npm",
			Rebuild:        []string{"npm ci"},
			Git: &GitMeta{
				Remote: "git@github.com:example/web-app.git",
				Branch: "main",
			},
		},
	}

	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid manifest, got: %v", err)
	}

	// Invalid manifest missing required backup_meta fields
	invalidM := NewManifest()
	if err := invalidM.Validate(); err == nil {
		t.Fatalf("expected validation error for empty backup_meta, got nil")
	}
}

func TestManifestYAMLAndPublicJSON(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "manifest.yaml")
	jsonPath := filepath.Join(tmpDir, "manifest.public.json")

	m := NewManifest()
	m.BackupMeta = BackupMeta{
		ID:               "test-id",
		Created:          time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC),
		CarryoverVersion: "1.0.0",
		Source: SourceOS{
			Hostname:       "host",
			OSID:           "fedora",
			OSVersion:      "40",
			Arch:           "x86_64",
			PackageManager: "dnf",
			Username:       "user",
			Home:           "/home/user",
			Shell:          "/bin/bash",
		},
	}
	m.Apps.Dnf = []PkgItem{{Name: "git"}}

	if err := m.SaveYAML(yamlPath); err != nil {
		t.Fatalf("SaveYAML failed: %v", err)
	}

	loaded, err := LoadYAML(yamlPath)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}
	if loaded.BackupMeta.Source.OSID != "fedora" {
		t.Errorf("expected os_id fedora, got %s", loaded.BackupMeta.Source.OSID)
	}

	if err := m.SavePublicJSON(jsonPath); err != nil {
		t.Fatalf("SavePublicJSON failed: %v", err)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("reading public json failed: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("public json was empty")
	}
}
