package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEndBackupAndRestoreCLI(t *testing.T) {
	binPath, err := filepath.Abs("../../bin/carryover")
	if err != nil {
		t.Fatalf("failed to resolve binary path: %v", err)
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Skipf("skipping CLI integration test: binary %s not built", binPath)
	}

	tmpDir := t.TempDir()
	sourceHome := filepath.Join(tmpDir, "source_home")
	destRepo := filepath.Join(tmpDir, "repo")
	targetHome := filepath.Join(tmpDir, "target_home")

	// 1. Seed source home
	_ = os.MkdirAll(filepath.Join(sourceHome, "Documents"), 0755)
	_ = os.WriteFile(filepath.Join(sourceHome, "Documents", "data.txt"), []byte("restored source data"), 0644)

	_ = os.MkdirAll(filepath.Join(sourceHome, ".ssh"), 0700)
	_ = os.WriteFile(filepath.Join(sourceHome, ".ssh", "id_ed25519"), []byte("secret_key"), 0600)
	_ = os.WriteFile(filepath.Join(sourceHome, ".ssh", "id_ed25519.pub"), []byte("pub_key"), 0644)

	_ = os.WriteFile(filepath.Join(sourceHome, ".bashrc"), []byte("alias mytool='run_tool'\n"), 0644)

	// Seed node project with excluded node_modules
	nodeDir := filepath.Join(sourceHome, "Projects", "app")
	nodeModules := filepath.Join(nodeDir, "node_modules", "pkg")
	_ = os.MkdirAll(nodeModules, 0755)
	_ = os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(nodeModules, "temp.js"), []byte("cache"), 0644)

	// 2. Perform Backup via CLI
	backupCmd := exec.Command(
		binPath,
		"backup",
		"--dest", destRepo,
		"--password", "secretpass123",
		"--home", sourceHome,
	)
	out, err := backupCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("carryover backup failed: %s (%v)", string(out), err)
	}

	// 3. Seed target home with existing conflict
	_ = os.MkdirAll(filepath.Join(targetHome, "Documents"), 0755)
	_ = os.WriteFile(filepath.Join(targetHome, "Documents", "data.txt"), []byte("original target data"), 0644)
	_ = os.WriteFile(filepath.Join(targetHome, ".bashrc"), []byte("# target default\n"), 0644)

	// 4. Perform Restore via CLI
	restoreCmd := exec.Command(
		binPath,
		"restore",
		"--dest", destRepo,
		"--password", "secretpass123",
		"--home", targetHome,
	)
	out, err = restoreCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("carryover restore failed: %s (%v)", string(out), err)
	}

	// 5. Assertions

	// Restored file placed
	restoredData, err := os.ReadFile(filepath.Join(targetHome, "Documents", "data.txt"))
	if err != nil || string(restoredData) != "restored source data" {
		t.Errorf("expected restored data in target home, got: %s (err: %v)", string(restoredData), err)
	}

	// Old file moved to ~/.carryover-old/
	carryoverOldDir := filepath.Join(targetHome, ".carryover-old")
	entries, err := os.ReadDir(carryoverOldDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected .carryover-old timestamp directory in %s", carryoverOldDir)
	}
	tsDir := filepath.Join(carryoverOldDir, entries[0].Name())
	oldBackup := filepath.Join(tsDir, "Documents", "data.txt")
	oldData, err := os.ReadFile(oldBackup)
	if err != nil || string(oldData) != "original target data" {
		t.Errorf("expected original data preserved in .carryover-old, got: %s (err: %v)", string(oldData), err)
	}

	// .bashrc merged
	bashrcBytes, _ := os.ReadFile(filepath.Join(targetHome, ".bashrc"))
	bashrcStr := string(bashrcBytes)
	if !strings.Contains(bashrcStr, "# target default") || !strings.Contains(bashrcStr, "alias mytool='run_tool'") {
		t.Errorf("expected merged .bashrc, got: %s", bashrcStr)
	}

	// SSH key permissions
	keyFi, err := os.Stat(filepath.Join(targetHome, ".ssh", "id_ed25519"))
	if err != nil || keyFi.Mode().Perm() != 0600 {
		t.Errorf("expected id_ed25519 mode 0600, got: %#o", keyFi.Mode().Perm())
	}
}
