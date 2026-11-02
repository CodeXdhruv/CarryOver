package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/manifest"
)

// BackupOptions defines parameters for a backup run
type BackupOptions struct {
	Dest           string
	Password       string
	PasswordFile   string
	Manifest       *manifest.Manifest
	DryRun         bool
	Emitter        *events.Emitter
	Runner         *ResticRunner
	BaseDir        string // defaults to user $HOME
}

// BackupResult summarizes the completed backup
type BackupResult struct {
	SnapshotID          string `json:"snapshot_id"`
	FilesNew            int64  `json:"files_new"`
	FilesChanged        int64  `json:"files_changed"`
	FilesUnmodified     int64  `json:"files_unmodified"`
	TotalBytesProcessed int64  `json:"total_bytes_processed"`
	ManifestSHA256      string `json:"manifest_sha256"`
}

// CheckDiskSpace returns available bytes on target filesystem path
func CheckDiskSpace(path string) (uint64, error) {
	// If path doesn't exist yet, check its parent
	for p := path; p != "" && p != "/"; p = filepath.Dir(p) {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(p, &stat); err == nil {
			return stat.Bavail * uint64(stat.Bsize), nil
		}
	}
	return 0, fmt.Errorf("unable to check disk space for %s", path)
}

// RunBackupPipeline executes the full backup workflow
func RunBackupPipeline(opts BackupOptions) (*BackupResult, error) {
	emitter := opts.Emitter
	if emitter == nil {
		emitter = events.NewEmitter(nil, false)
	}

	emitter.Phase("backup", "start")

	if opts.Manifest == nil {
		return nil, fmt.Errorf("manifest is required for backup")
	}

	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = opts.Manifest.BackupMeta.Source.Home
	}
	if baseDir == "" {
		var err error
		baseDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
	}

	// 1. Check destination writability and disk space
	dest := opts.Dest
	if dest == "" {
		return nil, fmt.Errorf("destination repository path cannot be empty")
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	availableBytes, err := CheckDiskSpace(dest)
	if err == nil {
		var estTotal int64
		for _, f := range opts.Manifest.Folders {
			estTotal += f.SizeBytes
		}
		for _, d := range opts.Manifest.Dotfiles {
			estTotal += d.SizeBytes
		}
		if estTotal > 0 && uint64(estTotal) > availableBytes {
			emitter.Warning("DISK_SPACE_LOW", fmt.Sprintf("Estimated backup size (%d MB) exceeds free space (%d MB)", estTotal/(1024*1024), availableBytes/(1024*1024)), dest)
		}
	}

	// 2. Prepare restic runner
	runner := opts.Runner
	if runner == nil {
		var err error
		runner, err = NewResticRunner("")
		if err != nil {
			return nil, err
		}
	}

	// 3. Prepare password file
	pwFile := opts.PasswordFile
	var cleanupPwFile bool
	if pwFile == "" {
		if opts.Password == "" {
			return nil, fmt.Errorf("either password or password-file must be provided")
		}
		tmpFile, err := os.CreateTemp("", "carryover-pw-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp password file: %w", err)
		}
		pwFile = tmpFile.Name()
		cleanupPwFile = true
		if _, err := tmpFile.WriteString(opts.Password); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(pwFile)
			return nil, err
		}
		_ = tmpFile.Close()
		_ = os.Chmod(pwFile, 0600)
	}
	if cleanupPwFile {
		defer func() {
			_ = os.Remove(pwFile)
		}()
	}

	// 4. Initialize restic repository if needed
	if err := runner.InitRepo(dest, pwFile); err != nil {
		return nil, err
	}

	// 5. Create temporary working directory for lists and manifest
	stagingDir, err := os.MkdirTemp("", "carryover-backup-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stagingDir)

	manifestPath := filepath.Join(stagingDir, "manifest.yaml")
	if err := opts.Manifest.SaveYAML(manifestPath); err != nil {
		return nil, fmt.Errorf("failed to write staging manifest: %w", err)
	}

	manifestBytes, _ := os.ReadFile(manifestPath)
	hash := sha256.Sum256(manifestBytes)
	opts.Manifest.Checksums.ManifestSHA256 = hex.EncodeToString(hash[:])
	_ = opts.Manifest.SaveYAML(manifestPath)

	// Build files-from list
	filesFromPath := filepath.Join(stagingDir, "files_from.txt")
	fFiles, err := os.Create(filesFromPath)
	if err != nil {
		return nil, err
	}

	// Include manifest.yaml in snapshot
	_, _ = fFiles.WriteString(manifestPath + "\n")

	// Include selected folders
	for _, folder := range opts.Manifest.Folders {
		full := filepath.Join(baseDir, folder.Path)
		if _, err := os.Stat(full); err == nil {
			_, _ = fFiles.WriteString(full + "\n")
		}
	}
	// Include selected dotfiles
	for _, dot := range opts.Manifest.Dotfiles {
		full := filepath.Join(baseDir, dot.Path)
		if _, err := os.Stat(full); err == nil {
			_, _ = fFiles.WriteString(full + "\n")
		}
	}
	_ = fFiles.Close()

	// Build excludes list
	excludeFilePath := filepath.Join(stagingDir, "excludes.txt")
	fExcludes, err := os.Create(excludeFilePath)
	if err != nil {
		return nil, err
	}
	for _, exc := range opts.Manifest.Exclude.Global {
		_, _ = fExcludes.WriteString(exc + "\n")
	}
	for projRel, list := range opts.Manifest.Exclude.PerProject {
		for _, exc := range list {
			fullExclude := filepath.Join(baseDir, projRel, exc)
			_, _ = fExcludes.WriteString(fullExclude + "\n")
		}
	}
	_ = fExcludes.Close()

	// 6. Run restic backup
	summary, err := runner.RunBackup(
		dest,
		pwFile,
		baseDir,
		filesFromPath,
		excludeFilePath,
		opts.DryRun,
		func(status ResticStatus) {
			current := ""
			if len(status.CurrentFiles) > 0 {
				current = status.CurrentFiles[0]
			}
			percent := int(status.PercentDone * 100)
			emitter.Progress("backup", percent, status.BytesDone, status.TotalBytes, current)
		},
	)
	if err != nil {
		return nil, err
	}

	// 7. Write manifest.public.json beside destination repository
	publicJsonPath := filepath.Join(dest, "manifest.public.json")
	_ = opts.Manifest.SavePublicJSON(publicJsonPath)

	// 8. Run repository verification (5% data subset check) if not dry run
	if !opts.DryRun {
		emitter.Progress("backup", 98, 0, 0, "Verifying backup integrity")
		if err := runner.VerifyRepo(dest, pwFile, 5); err != nil {
			emitter.Warning("VERIFY_WARN", fmt.Sprintf("Verification warning: %v", err), dest)
		}
	}

	result := &BackupResult{
		SnapshotID:          summary.SnapshotID,
		FilesNew:            summary.FilesNew,
		FilesChanged:        summary.FilesChanged,
		FilesUnmodified:     summary.FilesUnmodified,
		TotalBytesProcessed: summary.TotalBytesProcessed,
		ManifestSHA256:      opts.Manifest.Checksums.ManifestSHA256,
	}

	emitter.Done("ok", dest)
	emitter.Phase("backup", "complete")

	return result, nil
}
