package scanner

import (
	"fmt"
	"time"

	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/manifest"
)

// ScanOptions configures the scan operation
type ScanOptions struct {
	RootDir       string
	HomeDir       string
	CommandRunner CommandRunner
	Emitter       *events.Emitter
}

// ScanSystem runs the complete scan pipeline and returns an assembled manifest
func ScanSystem(opts ScanOptions) (*manifest.Manifest, error) {
	emitter := opts.Emitter
	if emitter == nil {
		emitter = events.NewEmitter(nil, false)
	}

	emitter.Phase("scan", "start")

	// 1. Detect OS
	emitter.Progress("scan", 10, 0, 0, "Detecting operating system")
	osInfo, err := DetectOS(opts.RootDir)
	if err != nil {
		// Fallback to reasonable defaults if /etc/os-release is missing
		osInfo = &OSInfo{
			ID:             "linux",
			PrettyName:     "Linux",
			PackageManager: "unknown",
		}
	}

	sysEnv := DetectSystemEnvironment()
	homeDir := opts.HomeDir
	if homeDir == "" {
		homeDir = sysEnv.Home
	}

	// 2. Discover apps
	emitter.Progress("scan", 30, 0, 0, "Cataloging installed applications")
	apps := ScanApps(opts.CommandRunner, osInfo.PackageManager)

	// 3. Discover folders
	emitter.Progress("scan", 55, 0, 0, "Discovering candidate folders")
	folders := ScanFolders(homeDir)

	// 4. Discover dotfiles
	emitter.Progress("scan", 70, 0, 0, "Scanning configuration and dotfiles")
	dotfiles := ScanDotfiles(homeDir)

	// 5. Capture shell configuration
	emitter.Progress("scan", 85, 0, 0, "Capturing shell configuration")
	shellConfig := ScanShellConfig(homeDir, sysEnv.Shell)

	// 6. Discover projects
	emitter.Progress("scan", 95, 0, 0, "Detecting developer projects")
	projectRoots := []string{"Projects", "dev", "code", "workspace", "src"}
	projects, perProjectExcludes := ScanProjects(homeDir, projectRoots, 4)

	// Assemble manifest
	m := manifest.NewManifest()
	m.BackupMeta = manifest.BackupMeta{
		ID:               time.Now().UTC().Format("2006-01-02T15-04-05Z"),
		Created:          time.Now().UTC(),
		CarryoverVersion: "1.0.0",
		Source: manifest.SourceOS{
			Hostname:       sysEnv.Hostname,
			OSID:           osInfo.ID,
			OSVersion:      osInfo.VersionID,
			Arch:           sysEnv.Arch,
			PackageManager: osInfo.PackageManager,
			Username:       sysEnv.Username,
			Home:           homeDir,
			Shell:          sysEnv.Shell,
			Desktop:        sysEnv.Desktop,
		},
	}

	m.Folders = folders
	m.Dotfiles = dotfiles
	m.Shell = shellConfig
	m.Apps = apps
	m.Projects = projects
	if len(perProjectExcludes) > 0 {
		m.Exclude.PerProject = perProjectExcludes
	}

	// Validate manifest against schema
	if err := m.Validate(); err != nil {
		emitter.Warning("MANIFEST_VALIDATION_WARN", fmt.Sprintf("Manifest validation warning: %v", err), "")
	}

	emitter.Phase("scan", "complete")
	return m, nil
}
