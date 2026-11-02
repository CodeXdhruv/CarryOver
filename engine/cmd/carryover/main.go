package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/events"
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

	// backup placeholder
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Perform backup to specified destination",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("backup engine will be initialized in milestone M2")
		},
	}
	backupCmd.Flags().String("dest", "", "Destination path or remote")
	backupCmd.Flags().String("select", "", "Path to selection.yaml")
	backupCmd.Flags().String("password-file", "", "Path to password file")
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

	// list placeholder
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots in backup destination",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("list will be initialized in milestone M2")
		},
	}
	listCmd.Flags().String("dest", "", "Backup destination")

	// verify placeholder
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify backup integrity and manifest checksum",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("verify will be initialized in milestone M2")
		},
	}
	verifyCmd.Flags().String("dest", "", "Backup destination")

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
