package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/carryover/carryover/engine/internal/backup"
	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/manifest"
	"github.com/carryover/carryover/engine/internal/mapper"
	"github.com/carryover/carryover/engine/internal/restore"
	"github.com/carryover/carryover/engine/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	jsonFlag bool
	outFlag  string
	rootFlag string
	homeFlag string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "carryover",
		Short: "Carryover: Linux OS Migration Application",
		Long:  "Carryover seamlessly migrates files, dotfiles, apps, and dev projects across Linux distros.",
	}

	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output newline-delimited JSON events")
	rootCmd.PersistentFlags().StringVar(&rootFlag, "root", "/", "Root filesystem path (useful for containers/testing)")
	rootCmd.PersistentFlags().StringVar(&homeFlag, "home", "", "User home directory path (useful for containers/testing)")

	// scan command
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the system for installed apps, configs, folders, and projects",
		RunE:  runScan,
	}
	scanCmd.Flags().StringVar(&outFlag, "out", "", "Output path for the scan manifest (e.g. scan.json or scan.yaml)")

	// backup command
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Perform backup to specified destination",
		RunE:  runBackup,
	}
	backupCmd.Flags().String("dest", "", "Destination path or remote")
	backupCmd.Flags().String("select", "", "Path to selection manifest / YAML")
	backupCmd.Flags().String("password-file", "", "Path to password file")
	backupCmd.Flags().String("password", "", "Backup encryption password")
	backupCmd.Flags().Bool("dry-run", false, "Simulate backup without writing data")

	// restore command
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore files, apps, dotfiles, and projects from a backup",
		RunE:  runRestore,
	}
	restoreCmd.Flags().String("dest", "", "Backup repository path")
	restoreCmd.Flags().String("snapshot", "latest", "Snapshot ID or latest")
	restoreCmd.Flags().String("password-file", "", "Path to password file")
	restoreCmd.Flags().String("password", "", "Backup encryption password")
	restoreCmd.Flags().Bool("plan-only", false, "Generate plan without restoring")
	restoreCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
	restoreCmd.Flags().String("only", "", "Comma-separated list of items to restore: files,dotfiles,apps,projects")
	restoreCmd.Flags().String("conflict", "keep_both", "Conflict policy: keep_both, overwrite, skip")

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots in backup destination",
		RunE:  runList,
	}
	listCmd.Flags().String("dest", "", "Backup destination")
	listCmd.Flags().String("password-file", "", "Path to password file")
	listCmd.Flags().String("password", "", "Backup encryption password")

	// verify command
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify backup integrity and manifest checksum",
		RunE:  runVerify,
	}
	verifyCmd.Flags().String("dest", "", "Backup destination")
	verifyCmd.Flags().String("password-file", "", "Path to password file")
	verifyCmd.Flags().String("password", "", "Backup encryption password")

	// map command
	mapCmd := &cobra.Command{
		Use:   "map <pkg>",
		Short: "Debug package name mapping between distributions",
		Args:  cobra.ExactArgs(1),
		RunE:  runMap,
	}
	mapCmd.Flags().String("from", "apt", "Source package manager")
	mapCmd.Flags().String("to", "pacman", "Target package manager")

	// export placeholder
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export migration recipe into other formats (e.g. Ansible)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("export will be initialized in milestone M8/M9")
		},
	}
	exportCmd.Flags().String("format", "ansible", "Export format")
	exportCmd.Flags().String("out", "playbook.yml", "Output file")

	// doctor placeholder
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system prerequisites (restic, privileges, disk space)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Carryover doctor check:")
			fmt.Printf("OS: %s\n", rootFlag)
			return nil
		},
	}

	rootCmd.AddCommand(scanCmd, backupCmd, restoreCmd, listCmd, verifyCmd, mapCmd, exportCmd, doctorCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	emitter := events.NewEmitter(os.Stdout, jsonFlag)

	opts := scanner.ScanOptions{
		RootDir: rootFlag,
		HomeDir: homeFlag,
		Emitter: emitter,
	}

	m, err := scanner.ScanSystem(opts)
	if err != nil {
		if jsonFlag {
			_ = emitter.Emit(events.Event{
				Type:    events.TypeError,
				Message: err.Error(),
			})
		}
		return err
	}

	// Write output file if requested
	if outFlag != "" {
		ext := strings.ToLower(filepath.Ext(outFlag))
		if ext == ".yaml" || ext == ".yml" {
			if err := m.SaveYAML(outFlag); err != nil {
				return fmt.Errorf("failed to save manifest YAML: %w", err)
			}
		} else {
			data, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(outFlag, data, 0644); err != nil {
				return fmt.Errorf("failed to save manifest JSON: %w", err)
			}
		}
	}

	if jsonFlag {
		// Output the completed manifest as JSON
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		// Also emit a final done event
		emitter.Done("ok", outFlag)
		// If no output file was specified, print the full JSON payload
		if outFlag == "" {
			fmt.Println(string(data))
		}
	} else {
		// Human readable CLI output
		fmt.Printf("Carryover Scan Complete\n")
		fmt.Printf("-----------------------\n")
		fmt.Printf("OS:              %s %s (%s)\n", m.BackupMeta.Source.OSID, m.BackupMeta.Source.OSVersion, m.BackupMeta.Source.PackageManager)
		fmt.Printf("Host:            %s (%s)\n", m.BackupMeta.Source.Hostname, m.BackupMeta.Source.Arch)
		fmt.Printf("User:            %s (%s)\n", m.BackupMeta.Source.Username, m.BackupMeta.Source.Home)
		fmt.Printf("Folders:         %d candidate directories found\n", len(m.Folders))
		for _, f := range m.Folders {
			fmt.Printf("  - %-20s (%d KB)\n", f.Path, f.SizeBytes/1024)
		}
		fmt.Printf("Dotfiles:        %d configuration files/dirs\n", len(m.Dotfiles))
		fmt.Printf("Shell:           %s (frameworks: %s)\n", m.Shell.Default, strings.Join(m.Shell.Frameworks, ", "))
		fmt.Printf("Applications:\n")
		if len(m.Apps.Apt) > 0 {
			fmt.Printf("  - apt:         %d packages\n", len(m.Apps.Apt))
		}
		if len(m.Apps.Dnf) > 0 {
			fmt.Printf("  - dnf:         %d packages\n", len(m.Apps.Dnf))
		}
		if len(m.Apps.Pacman) > 0 {
			fmt.Printf("  - pacman:      %d packages\n", len(m.Apps.Pacman))
		}
		if len(m.Apps.Zypper) > 0 {
			fmt.Printf("  - zypper:      %d packages\n", len(m.Apps.Zypper))
		}
		if len(m.Apps.Flatpak) > 0 {
			fmt.Printf("  - flatpak:     %d applications\n", len(m.Apps.Flatpak))
		}
		if len(m.Apps.Snap) > 0 {
			fmt.Printf("  - snap:        %d packages\n", len(m.Apps.Snap))
		}
		if len(m.Apps.NpmGlobal) > 0 {
			fmt.Printf("  - npm global:  %d packages\n", len(m.Apps.NpmGlobal))
		}
		if len(m.Apps.PipUser) > 0 {
			fmt.Printf("  - pip:         %d packages\n", len(m.Apps.PipUser))
		}
		if len(m.Apps.Cargo) > 0 {
			fmt.Printf("  - cargo:       %d packages\n", len(m.Apps.Cargo))
		}
		if len(m.Apps.VSCodeExtensions) > 0 {
			fmt.Printf("  - vscode:      %d extensions\n", len(m.Apps.VSCodeExtensions))
		}
		fmt.Printf("Projects:        %d detected projects\n", len(m.Projects))
		for _, p := range m.Projects {
			fmt.Printf("  - %-25s [%s] rebuild: %s\n", p.Path, p.Type, strings.Join(p.Rebuild, "; "))
		}
		if outFlag != "" {
			fmt.Printf("\nSaved manifest to: %s\n", outFlag)
		}
	}

	return nil
}

func runBackup(cmd *cobra.Command, args []string) error {
	dest, _ := cmd.Flags().GetString("dest")
	if dest == "" {
		return fmt.Errorf("--dest is required")
	}
	selectFile, _ := cmd.Flags().GetString("select")
	passwordFile, _ := cmd.Flags().GetString("password-file")
	password, _ := cmd.Flags().GetString("password")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if password == "" && passwordFile == "" {
		password = os.Getenv("RESTIC_PASSWORD")
	}
	if password == "" && passwordFile == "" {
		return fmt.Errorf("either --password, --password-file, or RESTIC_PASSWORD env var is required")
	}

	emitter := events.NewEmitter(os.Stdout, jsonFlag)

	// If selectFile is provided, load it; otherwise run system scan to get defaults
	var m *manifest.Manifest
	if selectFile != "" {
		var err error
		m, err = manifest.LoadYAML(selectFile)
		if err != nil {
			return fmt.Errorf("failed to load selection manifest: %w", err)
		}
	} else {
		var err error
		m, err = scanner.ScanSystem(scanner.ScanOptions{
			RootDir: rootFlag,
			HomeDir: homeFlag,
			Emitter: emitter,
		})
		if err != nil {
			return fmt.Errorf("failed to generate backup selection from scan: %w", err)
		}
	}

	res, err := backup.RunBackupPipeline(backup.BackupOptions{
		Dest:         dest,
		Password:     password,
		PasswordFile: passwordFile,
		Manifest:     m,
		DryRun:       dryRun,
		Emitter:      emitter,
		BaseDir:      homeFlag,
	})
	if err != nil {
		if jsonFlag {
			_ = emitter.Emit(events.Event{
				Type:    events.TypeError,
				Message: err.Error(),
			})
		}
		return err
	}

	if !jsonFlag {
		fmt.Println("Carryover Backup Complete")
		fmt.Println("-------------------------")
		fmt.Printf("Destination:     %s\n", dest)
		fmt.Printf("Snapshot ID:     %s\n", res.SnapshotID)
		fmt.Printf("New Files:       %d\n", res.FilesNew)
		fmt.Printf("Changed Files:   %d\n", res.FilesChanged)
		fmt.Printf("Total Bytes:     %d MB\n", res.TotalBytesProcessed/(1024*1024))
		fmt.Printf("Manifest SHA256: %s\n", res.ManifestSHA256)
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	dest, _ := cmd.Flags().GetString("dest")
	if dest == "" {
		return fmt.Errorf("--dest is required")
	}
	passwordFile, _ := cmd.Flags().GetString("password-file")
	password, _ := cmd.Flags().GetString("password")
	if password == "" && passwordFile == "" {
		password = os.Getenv("RESTIC_PASSWORD")
	}
	if password == "" && passwordFile == "" {
		return fmt.Errorf("either --password, --password-file, or RESTIC_PASSWORD env var is required")
	}

	runner, err := backup.NewResticRunner("")
	if err != nil {
		return err
	}

	pwFile := passwordFile
	if pwFile == "" {
		tmp, err := os.CreateTemp("", "carryover-pw-*")
		if err != nil {
			return err
		}
		pwFile = tmp.Name()
		defer os.Remove(pwFile)
		_, _ = tmp.WriteString(password)
		_ = tmp.Close()
	}

	snapshots, err := runner.ListSnapshots(dest, pwFile)
	if err != nil {
		return err
	}

	if jsonFlag {
		data, _ := json.Marshal(snapshots)
		fmt.Println(string(data))
	} else {
		fmt.Printf("Snapshots in %s:\n", dest)
		for _, s := range snapshots {
			fmt.Printf("ID: %s  Time: %s  Host: %s  Tags: %s\n", s.ShortID, s.Time.Format("2006-01-02 15:04:05"), s.Hostname, strings.Join(s.Tags, ","))
		}
	}
	return nil
}

func runVerify(cmd *cobra.Command, args []string) error {
	dest, _ := cmd.Flags().GetString("dest")
	if dest == "" {
		return fmt.Errorf("--dest is required")
	}
	passwordFile, _ := cmd.Flags().GetString("password-file")
	password, _ := cmd.Flags().GetString("password")
	if password == "" && passwordFile == "" {
		password = os.Getenv("RESTIC_PASSWORD")
	}
	if password == "" && passwordFile == "" {
		return fmt.Errorf("either --password, --password-file, or RESTIC_PASSWORD env var is required")
	}

	runner, err := backup.NewResticRunner("")
	if err != nil {
		return err
	}

	pwFile := passwordFile
	if pwFile == "" {
		tmp, err := os.CreateTemp("", "carryover-pw-*")
		if err != nil {
			return err
		}
		pwFile = tmp.Name()
		defer os.Remove(pwFile)
		_, _ = tmp.WriteString(password)
		_ = tmp.Close()
	}

	if err := runner.VerifyRepo(dest, pwFile, 10); err != nil {
		return err
	}

	if jsonFlag {
		emitter := events.NewEmitter(os.Stdout, true)
		emitter.Done("ok", dest)
	} else {
		fmt.Printf("Repository %s verified successfully (10%% data subset checked).\n", dest)
	}
	return nil
}

func runRestore(cmd *cobra.Command, args []string) error {
	dest, _ := cmd.Flags().GetString("dest")
	if dest == "" {
		return fmt.Errorf("--dest is required")
	}
	snapshotID, _ := cmd.Flags().GetString("snapshot")
	passwordFile, _ := cmd.Flags().GetString("password-file")
	password, _ := cmd.Flags().GetString("password")
	planOnly, _ := cmd.Flags().GetBool("plan-only")
	onlyFilter, _ := cmd.Flags().GetString("only")
	conflictStr, _ := cmd.Flags().GetString("conflict")

	if password == "" && passwordFile == "" {
		password = os.Getenv("RESTIC_PASSWORD")
	}
	if password == "" && passwordFile == "" {
		return fmt.Errorf("either --password, --password-file, or RESTIC_PASSWORD env var is required")
	}

	emitter := events.NewEmitter(os.Stdout, jsonFlag)

	runner, err := backup.NewResticRunner("")
	if err != nil {
		return err
	}

	pwFile := passwordFile
	var cleanupPw bool
	if pwFile == "" {
		tmp, err := os.CreateTemp("", "carryover-pw-*")
		if err != nil {
			return err
		}
		pwFile = tmp.Name()
		cleanupPw = true
		_, _ = tmp.WriteString(password)
		_ = tmp.Close()
	}
	if cleanupPw {
		defer os.Remove(pwFile)
	}

	// 1. Resolve snapshot ID if "latest"
	if snapshotID == "latest" {
		snapshots, err := runner.ListSnapshots(dest, pwFile)
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		if len(snapshots) == 0 {
			return fmt.Errorf("no snapshots found in repository %s", dest)
		}
		snapshotID = snapshots[len(snapshots)-1].ID
	}

	// 2. Extract snapshot to temporary staging directory
	stagingDir, err := os.MkdirTemp("", "carryover-restore-stage-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stagingDir)

	emitter.Progress("restore", 10, 0, 0, "Extracting snapshot metadata")
	if err := restore.RestoreSnapshotToStaging(runner, dest, snapshotID, pwFile, stagingDir); err != nil {
		return err
	}

	// 3. Locate manifest.yaml in stagingDir
	var manifestPath string
	_ = filepath.Walk(stagingDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && filepath.Base(path) == "manifest.yaml" {
			manifestPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if manifestPath == "" {
		return fmt.Errorf("manifest.yaml not found in snapshot %s", snapshotID)
	}

	m, err := manifest.LoadYAML(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest from snapshot: %w", err)
	}

	targetHome := homeFlag
	if targetHome == "" {
		targetHome, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	conflictNormalized := strings.ReplaceAll(strings.ToLower(conflictStr), "-", "_")
	policy := restore.ConflictPolicy(conflictNormalized)
	if policy != restore.ConflictKeepBoth && policy != restore.ConflictOverwrite && policy != restore.ConflictSkip {
		policy = restore.ConflictKeepBoth
	}
	timestamp := time.Now().UTC().Format("20060102T150405Z")

	// 4. Build restore plan
	plan, err := restore.BuildRestorePlan(m, snapshotID, targetHome, policy, true, timestamp, onlyFilter)
	if err != nil {
		return err
	}

	if planOnly {
		if jsonFlag {
			data, _ := json.Marshal(plan)
			fmt.Println(string(data))
		} else {
			fmt.Printf("Carryover Restore Plan (Dry Run)\n")
			fmt.Printf("--------------------------------\n")
			fmt.Printf("Snapshot ID:     %s\n", plan.SnapshotID)
			fmt.Printf("Source OS:       %s %s\n", plan.SourceOS.OSID, plan.SourceOS.OSVersion)
			fmt.Printf("Conflict Policy: %s\n", plan.ConflictPolicy)
			fmt.Printf("Planned Files:   %d\n", len(plan.Files))
			for _, f := range plan.Files {
				status := "NEW"
				if f.Conflict {
					status = fmt.Sprintf("CONFLICT (%s)", f.Action)
				}
				fmt.Printf("  - %-25s -> %s [%s]\n", f.RelPath, f.TargetPath, status)
			}
			fmt.Printf("Planned Dotfiles:%d\n", len(plan.Dotfiles))
			for _, d := range plan.Dotfiles {
				mode := "PLACE"
				if d.Merge {
					mode = "MERGE"
				}
				fmt.Printf("  - %-25s [%s]\n", d.RelPath, mode)
			}
		}
		return nil
	}

	// 5. Execute restore plan
	report, err := restore.ExecuteRestorePlan(plan, stagingDir, targetHome, emitter)
	if err != nil {
		return err
	}

	if jsonFlag {
		data, _ := json.Marshal(report)
		fmt.Println(string(data))
		emitter.Done("ok", "")
	} else {
		fmt.Println("Carryover Restore Complete")
		fmt.Println("--------------------------")
		fmt.Printf("Restored Files:  %d\n", len(report.RestoredFiles))
		fmt.Printf("Merged Dotfiles: %d\n", len(report.MergedDotfiles))
		fmt.Printf("Conflicts:       %d\n", len(report.ConflictedFiles))
		if len(report.ConflictedFiles) > 0 && report.OldBackupDir != "" {
			fmt.Printf("Old copies moved to: %s\n", report.OldBackupDir)
		}
		if len(report.FailedItems) > 0 {
			fmt.Printf("Failed Items:    %d\n", len(report.FailedItems))
		}
	}

	return nil
}

func runMap(cmd *cobra.Command, args []string) error {
	pkgName := args[0]
	fromPM, _ := cmd.Flags().GetString("from")
	toPM, _ := cmd.Flags().GetString("to")

	m := mapper.NewMapper()
	res := m.Resolve(pkgName, fromPM, toPM)

	if jsonFlag {
		data, _ := json.Marshal(res)
		fmt.Println(string(data))
	} else {
		fmt.Printf("Package Mapping Resolution:\n")
		fmt.Printf("---------------------------\n")
		fmt.Printf("Source:      %s (%s)\n", res.OriginalName, res.FromPM)
		fmt.Printf("Target:      %s (%s)\n", res.TargetName, res.ToPM)
		fmt.Printf("Confidence:  %s\n", res.Confidence)
		fmt.Printf("Auto-Select: %t\n", res.AutoSelect)
		fmt.Printf("Reason:      %s\n", res.Reason)
	}
	return nil
}
