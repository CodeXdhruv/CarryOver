package backup

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LocateRestic searches for the restic executable in bundled paths, then PATH
func LocateRestic() (string, error) {
	// 1. Check relative to current executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(dir, "restic"),
			filepath.Join(dir, "bin", "restic"),
			filepath.Join(dir, "..", "bin", "restic"),
		}
		for _, c := range candidates {
			if fi, err := os.Stat(c); err == nil && !fi.IsDir() && (fi.Mode()&0111 != 0) {
				return c, nil
			}
		}
	}

	// 2. Check ~/.local/bin/restic
	if home, err := os.UserHomeDir(); err == nil {
		localBin := filepath.Join(home, ".local", "bin", "restic")
		if fi, err := os.Stat(localBin); err == nil && !fi.IsDir() {
			return localBin, nil
		}
	}

	// 3. Fall back to PATH
	return exec.LookPath("restic")
}

// ResticStatus represents intermediate restic progress JSON
type ResticStatus struct {
	MessageType  string   `json:"message_type"`
	PercentDone  float64  `json:"percent_done"`
	TotalFiles   int64    `json:"total_files"`
	FilesDone    int64    `json:"files_done"`
	TotalBytes   int64    `json:"total_bytes"`
	BytesDone    int64    `json:"bytes_done"`
	CurrentFiles []string `json:"current_files"`
}

// ResticSummary represents the final restic summary JSON
type ResticSummary struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int64   `json:"files_new"`
	FilesChanged        int64   `json:"files_changed"`
	FilesUnmodified     int64   `json:"files_unmodified"`
	DataAdded           int64   `json:"data_added"`
	TotalFilesProcessed int64   `json:"total_files_processed"`
	TotalBytesProcessed int64   `json:"total_bytes_processed"`
	SnapshotID          string  `json:"snapshot_id"`
}

// Snapshot represents a restic snapshot metadata item
type Snapshot struct {
	ID       string    `json:"id"`
	ShortID  string    `json:"short_id"`
	Time     time.Time `json:"time"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	Tags     []string  `json:"tags"`
}

// ResticRunner handles interaction with restic binary
type ResticRunner struct {
	BinaryPath string
}

// NewResticRunner creates a restic runner with auto-located or given binary
func NewResticRunner(bin string) (*ResticRunner, error) {
	if bin == "" {
		var err error
		bin, err = LocateRestic()
		if err != nil {
			return nil, fmt.Errorf("restic binary not found: %w", err)
		}
	}
	return &ResticRunner{BinaryPath: bin}, nil
}

// InitRepo initializes restic repository at dest if not already initialized
func (r *ResticRunner) InitRepo(dest, passwordFile string) error {
	cmd := exec.Command(r.BinaryPath, "init", "--repo", dest)
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD_FILE="+passwordFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "config file already exists") {
			return nil
		}
		return fmt.Errorf("failed to init restic repo: %s (%w)", string(out), err)
	}
	return nil
}

// RunBackup executes restic backup and streams progress
func (r *ResticRunner) RunBackup(
	dest string,
	passwordFile string,
	baseDir string,
	filesFrom string,
	excludeFile string,
	dryRun bool,
	progressFn func(status ResticStatus),
) (*ResticSummary, error) {
	args := []string{
		"backup",
		"--repo", dest,
		"--files-from", filesFrom,
		"--exclude-file", excludeFile,
		"--tag", "carryover",
		"--json",
	}
	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command(r.BinaryPath, args...)
	cmd.Dir = baseDir
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD_FILE="+passwordFile)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start restic backup: %w", err)
	}

	var summary *ResticSummary
	scanner := bufio.NewScanner(stdout)
	lastProgressTime := time.Time{}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		msgType, _ := raw["message_type"].(string)
		if msgType == "status" {
			var st ResticStatus
			if err := json.Unmarshal(line, &st); err == nil && progressFn != nil {
				// Throttle progress to at most 4 updates per second
				now := time.Now()
				if now.Sub(lastProgressTime) >= 250*time.Millisecond || st.PercentDone >= 1.0 {
					lastProgressTime = now
					progressFn(st)
				}
			}
		} else if msgType == "summary" {
			var sm ResticSummary
			if err := json.Unmarshal(line, &sm); err == nil {
				summary = &sm
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("restic backup failed: %s (%w)", stderr.String(), err)
	}

	if summary == nil {
		summary = &ResticSummary{}
	}
	return summary, nil
}

// ListSnapshots lists snapshots in restic repo
func (r *ResticRunner) ListSnapshots(dest, passwordFile string) ([]Snapshot, error) {
	cmd := exec.Command(r.BinaryPath, "snapshots", "--repo", dest, "--json")
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD_FILE="+passwordFile)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	var snapshots []Snapshot
	if err := json.Unmarshal(out, &snapshots); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot json: %w", err)
	}
	return snapshots, nil
}

// VerifyRepo checks repository integrity and a data subset
func (r *ResticRunner) VerifyRepo(dest, passwordFile string, readSubsetPercent int) error {
	args := []string{"check", "--repo", dest}
	if readSubsetPercent > 0 {
		args = append(args, fmt.Sprintf("--read-data-subset=%d%%", readSubsetPercent))
	}

	cmd := exec.Command(r.BinaryPath, args...)
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD_FILE="+passwordFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic verification check failed: %s (%w)", string(out), err)
	}
	return nil
}
