package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/manifest"
)

// ConflictPolicy defines how existing destination files are handled
type ConflictPolicy string

const (
	ConflictKeepBoth  ConflictPolicy = "keep_both"
	ConflictOverwrite ConflictPolicy = "overwrite"
	ConflictSkip      ConflictPolicy = "skip"
)

// ActionType defines the action planned for a file
type ActionType string

const (
	ActionPlace   ActionType = "place"
	ActionBackup  ActionType = "backup_and_place"
	ActionMerge   ActionType = "merge"
	ActionSkip    ActionType = "skip"
)

// FilePlanItem details an individual file restore action
type FilePlanItem struct {
	RelPath    string     `json:"rel_path"`
	TargetPath string     `json:"target_path"`
	Conflict   bool       `json:"conflict"`
	Action     ActionType `json:"action"`
	BackupPath string     `json:"backup_path,omitempty"`
}

// DotfilePlanItem details a configuration dotfile restore action
type DotfilePlanItem struct {
	RelPath    string     `json:"rel_path"`
	TargetPath string     `json:"target_path"`
	Sensitive  bool       `json:"sensitive"`
	Conflict   bool       `json:"conflict"`
	Action     ActionType `json:"action"`
	Merge      bool       `json:"merge"`
}

// PermissionPlanItem specifies file mode hardening
type PermissionPlanItem struct {
	Path string      `json:"path"`
	Mode os.FileMode `json:"mode"`
}

// RestorePlan holds the dry run / planned actions before execution
type RestorePlan struct {
	SnapshotID     string               `json:"snapshot_id"`
	SourceOS       manifest.SourceOS    `json:"source_os"`
	ConflictPolicy ConflictPolicy       `json:"conflict_policy"`
	Timestamp      string               `json:"timestamp"`
	Files          []FilePlanItem       `json:"files"`
	Dotfiles       []DotfilePlanItem    `json:"dotfiles"`
	Permissions    []PermissionPlanItem `json:"permissions"`
}

// BuildRestorePlan constructs an actionable restore plan given a manifest and target home directory
func BuildRestorePlan(
	m *manifest.Manifest,
	snapshotID string,
	targetHome string,
	policy ConflictPolicy,
	dotfileMerge bool,
	timestamp string,
	onlyFilter string,
) (*RestorePlan, error) {
	if m == nil {
		return nil, fmt.Errorf("manifest is required to build restore plan")
	}
	if targetHome == "" {
		var err error
		targetHome, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine target home directory: %w", err)
		}
	}
	if policy == "" {
		policy = ConflictKeepBoth
	}

	filterMap := make(map[string]bool)
	if onlyFilter != "" {
		for _, part := range strings.Split(onlyFilter, ",") {
			filterMap[strings.TrimSpace(part)] = true
		}
	}
	includeAll := len(filterMap) == 0

	plan := &RestorePlan{
		SnapshotID:     snapshotID,
		SourceOS:       m.BackupMeta.Source,
		ConflictPolicy: policy,
		Timestamp:      timestamp,
	}

	backupBaseDir := filepath.Join(targetHome, ".carryover-old", timestamp)

	// 1. Plan Folders/Files
	if includeAll || filterMap["files"] {
		for _, folder := range m.Folders {
			targetPath := filepath.Join(targetHome, folder.Path)
			conflict := false
			if _, err := os.Stat(targetPath); err == nil {
				conflict = true
			}

			action := ActionPlace
			backupPath := ""
			if conflict {
				switch policy {
				case ConflictKeepBoth:
					action = ActionBackup
					backupPath = filepath.Join(backupBaseDir, folder.Path)
				case ConflictOverwrite:
					action = ActionPlace
				case ConflictSkip:
					action = ActionSkip
				}
			}

			plan.Files = append(plan.Files, FilePlanItem{
				RelPath:    folder.Path,
				TargetPath: targetPath,
				Conflict:   conflict,
				Action:     action,
				BackupPath: backupPath,
			})
		}
	}

	// 2. Plan Dotfiles
	if includeAll || filterMap["dotfiles"] {
		for _, dot := range m.Dotfiles {
			targetPath := filepath.Join(targetHome, dot.Path)
			conflict := false
			if _, err := os.Stat(targetPath); err == nil {
				conflict = true
			}

			isMergeCandidate := dotfileMerge && (dot.Path == ".bashrc" || dot.Path == ".zshrc" || dot.Path == ".profile")

			action := ActionPlace
			if conflict {
				if isMergeCandidate {
					action = ActionMerge
				} else {
					switch policy {
					case ConflictKeepBoth:
						action = ActionBackup
					case ConflictOverwrite:
						action = ActionPlace
					case ConflictSkip:
						action = ActionSkip
					}
				}
			}

			plan.Dotfiles = append(plan.Dotfiles, DotfilePlanItem{
				RelPath:    dot.Path,
				TargetPath: targetPath,
				Sensitive:  dot.Sensitive,
				Conflict:   conflict,
				Action:     action,
				Merge:      isMergeCandidate,
			})
		}
	}

	// 3. Plan SSH Permission Hardening
	sshDir := filepath.Join(targetHome, ".ssh")
	plan.Permissions = append(plan.Permissions, PermissionPlanItem{
		Path: sshDir,
		Mode: 0700,
	})

	return plan, nil
}
