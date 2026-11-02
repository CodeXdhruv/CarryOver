package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/manifest"
)

func TestBackupPipelineWithRestic(t *testing.T) {
	// Check if restic is available
	runner, err := NewResticRunner("")
	if err != nil {
		t.Skip("skipping restic integration test: restic not installed")
	}

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "fakehome")
	destRepo := filepath.Join(tmpDir, "backup_repo")

	// Seed files
	docsDir := filepath.Join(homeDir, "Documents")
	_ = os.MkdirAll(docsDir, 0755)
	_ = os.WriteFile(filepath.Join(docsDir, "report.txt"), []byte("Important report content"), 0644)

	// Seed node project with node_modules that should be excluded
	nodeProject := filepath.Join(homeDir, "Projects", "webapp")
	nodeModules := filepath.Join(nodeProject, "node_modules", "express")
	_ = os.MkdirAll(nodeModules, 0755)
	_ = os.WriteFile(filepath.Join(nodeProject, "package.json"), []byte(`{"name":"webapp"}`), 0644)
	_ = os.WriteFile(filepath.Join(nodeProject, "server.js"), []byte(`console.log("hello");`), 0644)
	_ = os.WriteFile(filepath.Join(nodeModules, "index.js"), []byte(`module.exports = {};`), 0644)

	// Create manifest
	m := manifest.NewManifest()
	m.BackupMeta = manifest.BackupMeta{
		ID:               "test-backup-001",
		Created:          time.Now().UTC(),
		CarryoverVersion: "1.0.0",
		Source: manifest.SourceOS{
			Hostname:       "testhost",
			OSID:           "ubuntu",
			OSVersion:      "24.04",
			Arch:           "x86_64",
			PackageManager: "apt",
			Username:       "testuser",
			Home:           homeDir,
			Shell:          "/bin/bash",
		},
	}
	m.Folders = []manifest.Folder{
		{Path: "Documents"},
		{Path: "Projects"},
	}
	m.Exclude.PerProject["Projects/webapp"] = []string{"node_modules"}

	var eventBuf bytes.Buffer
	emitter := events.NewEmitter(&eventBuf, true)

	opts := BackupOptions{
		Dest:         destRepo,
		Password:     "testpassword123",
		Manifest:     m,
		DryRun:       false,
		Emitter:      emitter,
		Runner:       runner,
		BaseDir:      homeDir,
	}

	result, err := RunBackupPipeline(opts)
	if err != nil {
		t.Fatalf("RunBackupPipeline failed: %v", err)
	}

	if result.SnapshotID == "" {
		t.Errorf("expected non-empty snapshot ID")
	}

	// Verify manifest.public.json written to destination
	publicPath := filepath.Join(destRepo, "manifest.public.json")
	if _, err := os.Stat(publicPath); err != nil {
		t.Errorf("expected manifest.public.json in dest repo: %v", err)
	}

	// Verify snapshots list
	pwFile := filepath.Join(tmpDir, "pw.txt")
	_ = os.WriteFile(pwFile, []byte("testpassword123"), 0600)

	snapshots, err := runner.ListSnapshots(destRepo, pwFile)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	}

	// Verify repository integrity check
	if err := runner.VerifyRepo(destRepo, pwFile, 100); err != nil {
		t.Errorf("VerifyRepo failed: %v", err)
	}

	// Verify events were emitted
	if eventBuf.Len() == 0 {
		t.Errorf("expected backup events on emitter stream")
	}
}
