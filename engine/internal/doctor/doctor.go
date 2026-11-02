package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/carryover/carryover/engine/internal/backup"
)

// CheckStatus represents the health status of a check
type CheckStatus string

const (
	StatusPass CheckStatus = "pass"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
)

// CheckResult represents a single diagnostic result
type CheckResult struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
	Details string      `json:"details,omitempty"`
}

// DoctorReport contains all diagnostic results
type DoctorReport struct {
	OverallStatus CheckStatus   `json:"overall_status"`
	Checks        []CheckResult `json:"checks"`
}

// RunDiagnostics executes system health checks
func RunDiagnostics(targetHome string) (*DoctorReport, error) {
	if targetHome == "" {
		var err error
		targetHome, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to determine home directory: %w", err)
		}
	}

	report := &DoctorReport{
		OverallStatus: StatusPass,
		Checks:        make([]CheckResult, 0),
	}

	addCheck := func(name string, status CheckStatus, message string, details string) {
		if status == StatusFail {
			report.OverallStatus = StatusFail
		} else if status == StatusWarn && report.OverallStatus != StatusFail {
			report.OverallStatus = StatusWarn
		}
		report.Checks = append(report.Checks, CheckResult{
			Name:    name,
			Status:  status,
			Message: message,
			Details: details,
		})
	}

	// 1. Restic binary check
	resticPath, err := backup.LocateRestic()
	if err != nil {
		addCheck("restic_binary", StatusFail, "Restic binary not found", "Install restic or place it in bin/restic")
	} else {
		cmd := exec.Command(resticPath, "version")
		out, err := cmd.Output()
		if err != nil {
			addCheck("restic_binary", StatusWarn, "Restic binary found but failed to report version", string(out))
		} else {
			addCheck("restic_binary", StatusPass, "Restic binary available", string(out))
		}
	}

	// 2. Privilege elevation tools check
	_, pkexecErr := exec.LookPath("pkexec")
	_, sudoErr := exec.LookPath("sudo")
	if pkexecErr == nil {
		addCheck("privilege_elevation", StatusPass, "pkexec is available for GUI elevation", "")
	} else if sudoErr == nil {
		addCheck("privilege_elevation", StatusWarn, "pkexec not found, falling back to CLI sudo", "")
	} else {
		addCheck("privilege_elevation", StatusFail, "Neither pkexec nor sudo found", "System package installation will require manual elevation")
	}

	// 3. Disk space check
	var stat syscall.Statfs_t
	if err := syscall.Statfs(targetHome, &stat); err != nil {
		addCheck("disk_space", StatusWarn, "Unable to inspect home filesystem disk space", err.Error())
	} else {
		freeBytes := stat.Bavail * uint64(stat.Bsize)
		freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
		if freeGB < 1.0 {
			addCheck("disk_space", StatusFail, fmt.Sprintf("Critically low disk space: %.1f GB available", freeGB), "Free up space before backup or restore")
		} else if freeGB < 5.0 {
			addCheck("disk_space", StatusWarn, fmt.Sprintf("Low disk space warning: %.1f GB available", freeGB), "")
		} else {
			addCheck("disk_space", StatusPass, fmt.Sprintf("Disk space adequate: %.1f GB available in %s", freeGB, targetHome), "")
		}
	}

	// 4. Dotfile and SSH permissions check
	sshDir := filepath.Join(targetHome, ".ssh")
	if fi, err := os.Stat(sshDir); err == nil {
		mode := fi.Mode().Perm()
		if mode != 0700 {
			addCheck("ssh_permissions", StatusWarn, fmt.Sprintf("~/.ssh directory permission is %04o (expected 0700)", mode), "Restore pipeline will harden this automatically")
		} else {
			addCheck("ssh_permissions", StatusPass, "~/.ssh directory permissions are secure (0700)", "")
		}
	} else {
		addCheck("ssh_permissions", StatusPass, "No ~/.ssh directory found", "")
	}

	// 5. Package manager check
	pms := []string{"apt-get", "dnf", "pacman", "zypper"}
	foundPM := false
	for _, pm := range pms {
		if _, err := exec.LookPath(pm); err == nil {
			addCheck("package_manager", StatusPass, fmt.Sprintf("Detected native package manager: %s", pm), "")
			foundPM = true
			break
		}
	}
	if !foundPM {
		addCheck("package_manager", StatusWarn, "No primary system package manager (apt/dnf/pacman/zypper) in PATH", "")
	}

	return report, nil
}
