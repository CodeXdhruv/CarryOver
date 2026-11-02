package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/backup"
	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/manifest"
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

	// scan command
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the system for installed apps, configs, folders, and projects",
		RunE:  runScan,
	}
	scanCmd.Flags().StringVar(&outFlag, "out", "", "Output path for the scan manifest (e.g. scan.json or scan.yaml)")
	scanCmd.Flags().StringVar(&rootFlag, "root", "/", "Root filesystem path for scanning (useful for containers/testing)")
	scanCmd.Flags().StringVar(&homeFlag, "home", "", "User home directory path for scanning")

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

	// restore placeholder
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore files, apps, dotfiles, and projects from a backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("restore engine will be initialized in milestone M3")
		},
	}
	restoreCmd.Flags().String("dest", "", "Backup repository path")
	restoreCmd.Flags().String("snapshot", "latest", "Snapshot ID or latest")
	restoreCmd.Flags().Bool("plan-only", false, "Generate plan without restoring")
	restoreCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
	restoreCmd.Flags().String("only", "", "Comma-separated list of items to restore: files,dotfiles,apps,projects")

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

	// map placeholder
	mapCmd := &cobra.Command{
		Use:   "map <pkg>",
		Short: "Debug package name mapping between distributions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("package mapper will be initialized in milestone M4")
		},
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
