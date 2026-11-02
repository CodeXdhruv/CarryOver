package restore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/backup"
	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/rebuild"
)

// RestoreReport summarizes the outcome of executing a restore plan
type RestoreReport struct {
	RestoredFiles   []string `json:"restored_files"`
	ConflictedFiles []string `json:"conflicted_files"`
	SkippedFiles    []string `json:"skipped_files"`
	MergedDotfiles  []string `json:"merged_dotfiles"`
	FailedItems     []string `json:"failed_items"`
	OldBackupDir    string   `json:"old_backup_dir,omitempty"`
}

// RestoreSnapshotToStaging calls restic restore to extract files to a temporary staging folder
func RestoreSnapshotToStaging(
	runner *backup.ResticRunner,
	dest string,
	snapshotID string,
	passwordFile string,
	stagingDir string,
) error {
	cmd := exec.Command(
		runner.BinaryPath,
		"restore",
		snapshotID,
		"--target", stagingDir,
		"--repo", dest,
	)
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD_FILE="+passwordFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic restore to staging failed: %s (%w)", string(out), err)
	}
	return nil
}

// MergeDotfileContent safely merges restored content into existing dotfile using carryover marker blocks
func MergeDotfileContent(targetPath string, restoredContent []byte) error {
	existingData, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(targetPath, restoredContent, 0644)
		}
		return err
	}

	startMarker := "# >>> carryover restored >>>"
	endMarker := "# <<< carryover restored <<<"

	contentStr := string(existingData)
	if strings.Contains(contentStr, startMarker) {
		// Marker already present; replace existing block or leave intact to remain idempotent
		startIndex := strings.Index(contentStr, startMarker)
		endIndex := strings.Index(contentStr, endMarker)
		if endIndex > startIndex {
			before := contentStr[:startIndex]
			after := contentStr[endIndex+len(endMarker):]
			newBlock := fmt.Sprintf("%s\n%s\n%s", startMarker, string(restoredContent), endMarker)
			return os.WriteFile(targetPath, []byte(before+newBlock+after), 0644)
		}
	}

	// Append new marker block
	var buf bytes.Buffer
	buf.Write(existingData)
	if !strings.HasSuffix(contentStr, "\n") {
		buf.WriteString("\n")
	}
	buf.WriteString("\n" + startMarker + "\n")
	buf.Write(restoredContent)
	if !bytes.HasSuffix(restoredContent, []byte("\n")) {
		buf.WriteString("\n")
	}
	buf.WriteString(endMarker + "\n")

	return os.WriteFile(targetPath, buf.Bytes(), 0644)
}

// CopyFile copies a regular file preserving permissions
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err == nil {
		_ = os.Chmod(dst, si.Mode())
	}
	return nil
}

// CopyDir recursively copies directories
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return CopyFile(path, target)
	})
}

// MoveToOldBackup moves an existing file or directory to ~/.carryover-old/<timestamp>/
func MoveToOldBackup(existingPath, targetHome, timestamp, relPath string) (string, error) {
	backupDir := filepath.Join(targetHome, ".carryover-old", timestamp, filepath.Dir(relPath))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}
	backupPath := filepath.Join(targetHome, ".carryover-old", timestamp, relPath)
	if err := os.Rename(existingPath, backupPath); err != nil {
		// Fallback to copy and remove if rename across devices fails
		if fi, err := os.Stat(existingPath); err == nil && fi.IsDir() {
			if err := CopyDir(existingPath, backupPath); err != nil {
				return "", err
			}
			_ = os.RemoveAll(existingPath)
		} else {
			if err := CopyFile(existingPath, backupPath); err != nil {
				return "", err
			}
			_ = os.Remove(existingPath)
		}
	}
	return backupPath, nil
}

// FixSSHPermissions ensures ~/.ssh is 0700 and private keys are 0600
func FixSSHPermissions(sshDir string) error {
	fi, err := os.Stat(sshDir)
	if err != nil {
		return nil
	}
	if fi.IsDir() {
		_ = os.Chmod(sshDir, 0700)
		entries, _ := os.ReadDir(sshDir)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			fullPath := filepath.Join(sshDir, name)
			// Public keys stay 0644; private keys hardened to 0600
			if strings.HasSuffix(name, ".pub") || name == "known_hosts" || name == "config" {
				_ = os.Chmod(fullPath, 0644)
			} else if strings.HasPrefix(name, "id_") || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") {
				_ = os.Chmod(fullPath, 0600)
			}
		}
	}
	return nil
}

// ExecuteRestorePlan performs the concrete file placement, merging, and hardening
func ExecuteRestorePlan(
	plan *RestorePlan,
	stagingDir string,
	targetHome string,
	emitter *events.Emitter,
) (*RestoreReport, error) {
	if emitter == nil {
		emitter = events.NewEmitter(nil, false)
	}

	emitter.Phase("restore_files", "start")
	report := &RestoreReport{
		OldBackupDir: filepath.Join(targetHome, ".carryover-old", plan.Timestamp),
	}

	findStagedPath := func(rel string) string {
		// Candidate 1: relative to stagingDir directly
		c1 := filepath.Join(stagingDir, rel)
		if _, err := os.Stat(c1); err == nil {
			return c1
		}
		// Candidate 2: inside source home tree under stagingDir
		if plan.SourceOS.Home != "" {
			c2 := filepath.Join(stagingDir, strings.TrimPrefix(plan.SourceOS.Home, "/"), rel)
			if _, err := os.Stat(c2); err == nil {
				return c2
			}
		}
		// Candidate 3: search stagingDir for basename
		var match string
		_ = filepath.Walk(stagingDir, func(p string, info os.FileInfo, err error) error {
			if err == nil && strings.HasSuffix(p, rel) {
				match = p
				return filepath.SkipAll
			}
			return nil
		})
		return match
	}

	// 1. Process files and folders
	for i, item := range plan.Files {
		emitter.Progress("restore_files", int(float64(i+1)/float64(len(plan.Files)+len(plan.Dotfiles))*100), 0, 0, item.RelPath)
		stagedPath := findStagedPath(item.RelPath)
		if stagedPath == "" {
			continue
		}

		if item.Conflict {
			report.ConflictedFiles = append(report.ConflictedFiles, item.RelPath)
			if item.Action == ActionSkip {
				report.SkippedFiles = append(report.SkippedFiles, item.RelPath)
				continue
			}
			if item.Action == ActionBackup {
				_, _ = MoveToOldBackup(item.TargetPath, targetHome, plan.Timestamp, item.RelPath)
			}
		}

		fi, err := os.Stat(stagedPath)
		if err != nil {
			report.FailedItems = append(report.FailedItems, item.RelPath)
			continue
		}

		if fi.IsDir() {
			if err := CopyDir(stagedPath, item.TargetPath); err != nil {
				report.FailedItems = append(report.FailedItems, item.RelPath)
			} else {
				report.RestoredFiles = append(report.RestoredFiles, item.RelPath)
			}
		} else {
			if err := CopyFile(stagedPath, item.TargetPath); err != nil {
				report.FailedItems = append(report.FailedItems, item.RelPath)
			} else {
				report.RestoredFiles = append(report.RestoredFiles, item.RelPath)
			}
		}
	}

	// 2. Process dotfiles
	for _, dot := range plan.Dotfiles {
		stagedPath := findStagedPath(dot.RelPath)
		if stagedPath == "" {
			continue
		}

		if dot.Merge {
			restoredBytes, err := os.ReadFile(stagedPath)
			if err == nil {
				if err := MergeDotfileContent(dot.TargetPath, restoredBytes); err == nil {
					report.MergedDotfiles = append(report.MergedDotfiles, dot.RelPath)
					continue
				}
			}
		}

		if dot.Conflict {
			report.ConflictedFiles = append(report.ConflictedFiles, dot.RelPath)
			if dot.Action == ActionSkip {
				report.SkippedFiles = append(report.SkippedFiles, dot.RelPath)
				continue
			}
			if dot.Action == ActionBackup {
				_, _ = MoveToOldBackup(dot.TargetPath, targetHome, plan.Timestamp, dot.RelPath)
			}
		}

		fi, err := os.Stat(stagedPath)
		if err != nil {
			report.FailedItems = append(report.FailedItems, dot.RelPath)
			continue
		}

		if fi.IsDir() {
			if err := CopyDir(stagedPath, dot.TargetPath); err != nil {
				report.FailedItems = append(report.FailedItems, dot.RelPath)
			} else {
				report.RestoredFiles = append(report.RestoredFiles, dot.RelPath)
			}
		} else {
			if err := CopyFile(stagedPath, dot.TargetPath); err != nil {
				report.FailedItems = append(report.FailedItems, dot.RelPath)
			} else {
				report.RestoredFiles = append(report.RestoredFiles, dot.RelPath)
			}
		}
	}

	// 3. Fix permissions
	_ = FixSSHPermissions(filepath.Join(targetHome, ".ssh"))

	// 4. Reinstall shell frameworks before final dotfile usage
	if len(plan.ShellFrameworks) > 0 {
		emitter.Phase("reinstall_frameworks", "start")
		rebuild.ReinstallAllFrameworks(targetHome, plan.ShellFrameworks, emitter)
		emitter.Phase("reinstall_frameworks", "complete")
	}

	// 5. Rebuild projects
	if len(plan.Projects) > 0 {
		emitter.Phase("rebuild_projects", "start")
		for _, proj := range plan.Projects {
			safeCmds := rebuild.DeriveSafeCommands(proj.Type, proj.PackageManager)
			if len(safeCmds) == 0 {
				safeCmds = proj.RebuildCommands
			}
			for _, cmd := range safeCmds {
				_ = rebuild.RunRebuild(rebuild.RebuildOptions{
					ProjectPath: proj.TargetPath,
					ProjectType: proj.Type,
					Command:     cmd,
					Emitter:     emitter,
				})
			}
		}
		emitter.Phase("rebuild_projects", "complete")
	}

	emitter.Phase("restore_files", "complete")
	return report, nil
}

// SaveReport writes report.json to specified file
func (r *RestoreReport) SaveReport(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
