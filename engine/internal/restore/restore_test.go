package restore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/carryover/carryover/engine/internal/manifest"
)

func TestBuildRestorePlan(t *testing.T) {
	tmpHome := t.TempDir()

	// Existing file creates conflict
	_ = os.MkdirAll(filepath.Join(tmpHome, "Documents"), 0755)
	_ = os.WriteFile(filepath.Join(tmpHome, "Documents", "existing.txt"), []byte("old"), 0644)
	_ = os.WriteFile(filepath.Join(tmpHome, ".bashrc"), []byte("# distro default"), 0644)

	m := manifest.NewManifest()
	m.Folders = []manifest.Folder{{Path: "Documents"}}
	m.Dotfiles = []manifest.Dotfile{
		{Path: ".bashrc"},
		{Path: ".gitconfig"},
	}

	plan, err := BuildRestorePlan(m, "snap123", tmpHome, ConflictKeepBoth, true, "20261102T120000Z", "")
	if err != nil {
		t.Fatalf("BuildRestorePlan failed: %v", err)
	}

	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file plan item, got %d", len(plan.Files))
	}
	if !plan.Files[0].Conflict {
		t.Errorf("expected conflict on Documents folder")
	}
	if plan.Files[0].Action != ActionBackup {
		t.Errorf("expected ActionBackup on conflict with KeepBoth policy")
	}

	// Dotfile merge check
	var bashrcPlan *DotfilePlanItem
	for i := range plan.Dotfiles {
		if plan.Dotfiles[i].RelPath == ".bashrc" {
			bashrcPlan = &plan.Dotfiles[i]
		}
	}
	if bashrcPlan == nil || !bashrcPlan.Merge || bashrcPlan.Action != ActionMerge {
		t.Errorf("expected .bashrc to be marked for merge, got: %+v", bashrcPlan)
	}
}

func TestMergeDotfileContentIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	bashrc := filepath.Join(tmpDir, ".bashrc")
	_ = os.WriteFile(bashrc, []byte("# distro default\nalias ls='ls --color'\n"), 0644)

	restoredSnippet := []byte("export EDITOR=nvim\nalias ll='ls -l'\n")

	// 1. First merge
	if err := MergeDotfileContent(bashrc, restoredSnippet); err != nil {
		t.Fatalf("first MergeDotfileContent failed: %v", err)
	}
	data1, _ := os.ReadFile(bashrc)
	str1 := string(data1)
	if !strings.Contains(str1, "# >>> carryover restored >>>") {
		t.Errorf("expected start marker in merged file")
	}
	if !strings.Contains(str1, "export EDITOR=nvim") {
		t.Errorf("expected restored content in merged file")
	}
	if strings.Count(str1, "# >>> carryover restored >>>") != 1 {
		t.Errorf("expected exactly 1 marker block")
	}

	// 2. Second merge (idempotent replay)
	if err := MergeDotfileContent(bashrc, restoredSnippet); err != nil {
		t.Fatalf("second MergeDotfileContent failed: %v", err)
	}
	data2, _ := os.ReadFile(bashrc)
	str2 := string(data2)
	if strings.Count(str2, "# >>> carryover restored >>>") != 1 {
		t.Errorf("expected exactly 1 marker block after replay, got %d", strings.Count(str2, "# >>> carryover restored >>>"))
	}
}

func TestMoveToOldBackup(t *testing.T) {
	tmpHome := t.TempDir()
	existingFile := filepath.Join(tmpHome, "Documents", "data.txt")
	_ = os.MkdirAll(filepath.Dir(existingFile), 0755)
	_ = os.WriteFile(existingFile, []byte("original data"), 0644)

	backupPath, err := MoveToOldBackup(existingFile, tmpHome, "20261102T120000Z", "Documents/data.txt")
	if err != nil {
		t.Fatalf("MoveToOldBackup failed: %v", err)
	}

	// Verify original moved
	if _, err := os.Stat(existingFile); !os.IsNotExist(err) {
		t.Errorf("expected original file to be moved")
	}

	// Verify backup copy exists
	backedData, err := os.ReadFile(backupPath)
	if err != nil || string(backedData) != "original data" {
		t.Errorf("unexpected backup file content: %s (err: %v)", string(backedData), err)
	}
}

func TestFixSSHPermissions(t *testing.T) {
	tmpHome := t.TempDir()
	sshDir := filepath.Join(tmpHome, ".ssh")
	_ = os.MkdirAll(sshDir, 0777)

	privKey := filepath.Join(sshDir, "id_ed25519")
	pubKey := filepath.Join(sshDir, "id_ed25519.pub")
	_ = os.WriteFile(privKey, []byte("private key"), 0666)
	_ = os.WriteFile(pubKey, []byte("public key"), 0666)

	if err := FixSSHPermissions(sshDir); err != nil {
		t.Fatalf("FixSSHPermissions failed: %v", err)
	}

	dirFi, _ := os.Stat(sshDir)
	if dirFi.Mode().Perm() != 0700 {
		t.Errorf("expected 0700 for .ssh dir, got %#o", dirFi.Mode().Perm())
	}

	privFi, _ := os.Stat(privKey)
	if privFi.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 for private key, got %#o", privFi.Mode().Perm())
	}

	pubFi, _ := os.Stat(pubKey)
	if pubFi.Mode().Perm() != 0644 {
		t.Errorf("expected 0644 for public key, got %#o", pubFi.Mode().Perm())
	}
}

func TestExecuteRestorePlan(t *testing.T) {
	tmpDir := t.TempDir()
	targetHome := filepath.Join(tmpDir, "newhome")
	stagingDir := filepath.Join(tmpDir, "staged")

	// Existing target files
	_ = os.MkdirAll(filepath.Join(targetHome, "Documents"), 0755)
	_ = os.WriteFile(filepath.Join(targetHome, "Documents", "notes.txt"), []byte("old notes"), 0644)
	_ = os.WriteFile(filepath.Join(targetHome, ".bashrc"), []byte("# ubuntu default\n"), 0644)

	// Staged files from backup
	_ = os.MkdirAll(filepath.Join(stagingDir, "Documents"), 0755)
	_ = os.WriteFile(filepath.Join(stagingDir, "Documents", "notes.txt"), []byte("restored notes"), 0644)
	_ = os.WriteFile(filepath.Join(stagingDir, ".bashrc"), []byte("alias ll='ls -l'\n"), 0644)

	m := manifest.NewManifest()
	m.Folders = []manifest.Folder{{Path: "Documents"}}
	m.Dotfiles = []manifest.Dotfile{{Path: ".bashrc"}}

	ts := time.Now().UTC().Format("20060102T150405Z")
	plan, err := BuildRestorePlan(m, "test-snap", targetHome, ConflictKeepBoth, true, ts, "")
	if err != nil {
		t.Fatalf("BuildRestorePlan failed: %v", err)
	}

	report, err := ExecuteRestorePlan(plan, stagingDir, targetHome, nil)
	if err != nil {
		t.Fatalf("ExecuteRestorePlan failed: %v", err)
	}

	// Verify old file backed up
	oldBackupFile := filepath.Join(targetHome, ".carryover-old", ts, "Documents", "notes.txt")
	if data, err := os.ReadFile(oldBackupFile); err != nil || string(data) != "old notes" {
		t.Errorf("expected old notes preserved in .carryover-old: %v", err)
	}

	// Verify restored file placed
	restoredFile := filepath.Join(targetHome, "Documents", "notes.txt")
	if data, err := os.ReadFile(restoredFile); err != nil || string(data) != "restored notes" {
		t.Errorf("expected restored notes placed: %v", err)
	}

	// Verify .bashrc merged
	bashrcData, _ := os.ReadFile(filepath.Join(targetHome, ".bashrc"))
	bashrcStr := string(bashrcData)
	if !strings.Contains(bashrcStr, "# ubuntu default") || !strings.Contains(bashrcStr, "alias ll='ls -l'") {
		t.Errorf("expected both distro default and restored content in .bashrc: %s", bashrcStr)
	}

	if len(report.ConflictedFiles) == 0 {
		t.Errorf("expected conflict reported")
	}
}
