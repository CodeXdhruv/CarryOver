package manifest

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed schema.json
var embeddedSchema []byte

var compiledSchema *jsonschema.Schema
var schemaErr error

func init() {
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", bytes.NewReader(embeddedSchema)); err != nil {
		schemaErr = err
		return
	}
	compiledSchema, schemaErr = c.Compile("schema.json")
}

// Manifest represents the top-level carryover manifest
type Manifest struct {
	SchemaVersion int         `json:"schema_version" yaml:"schema_version"`
	BackupMeta    BackupMeta  `json:"backup_meta" yaml:"backup_meta"`
	Folders       []Folder    `json:"folders,omitempty" yaml:"folders,omitempty"`
	Exclude       Exclude     `json:"exclude,omitempty" yaml:"exclude,omitempty"`
	Dotfiles      []Dotfile   `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
	Shell         ShellConfig `json:"shell,omitempty" yaml:"shell,omitempty"`
	Apps          AppSet      `json:"apps,omitempty" yaml:"apps,omitempty"`
	Projects      []Project   `json:"projects,omitempty" yaml:"projects,omitempty"`
	Checksums     Checksums   `json:"checksums,omitempty" yaml:"checksums,omitempty"`
}

type BackupMeta struct {
	ID               string    `json:"id" yaml:"id"`
	Created          time.Time `json:"created" yaml:"created"`
	CarryoverVersion string    `json:"carryover_version" yaml:"carryover_version"`
	Source           SourceOS  `json:"source" yaml:"source"`
}

type SourceOS struct {
	Hostname       string `json:"hostname" yaml:"hostname"`
	OSID           string `json:"os_id" yaml:"os_id"`
	OSVersion      string `json:"os_version" yaml:"os_version"`
	Arch           string `json:"arch" yaml:"arch"`
	PackageManager string `json:"package_manager" yaml:"package_manager"`
	Username       string `json:"username" yaml:"username"`
	Home           string `json:"home" yaml:"home"`
	Shell          string `json:"shell" yaml:"shell"`
	Desktop        string `json:"desktop,omitempty" yaml:"desktop,omitempty"`
}

type Folder struct {
	Path      string `json:"path" yaml:"path"`
	SizeBytes int64  `json:"size_bytes,omitempty" yaml:"size_bytes,omitempty"`
}

type Exclude struct {
	Global     []string            `json:"global,omitempty" yaml:"global,omitempty"`
	PerProject map[string][]string `json:"per_project,omitempty" yaml:"per_project,omitempty"`
}

type Dotfile struct {
	Path      string `json:"path" yaml:"path"`
	Sensitive bool   `json:"sensitive,omitempty" yaml:"sensitive,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty" yaml:"size_bytes,omitempty"`
}

type ShellConfig struct {
	Default       string            `json:"default,omitempty" yaml:"default,omitempty"`
	Frameworks    []string          `json:"frameworks,omitempty" yaml:"frameworks,omitempty"`
	EnvExports    map[string]string `json:"env_exports,omitempty" yaml:"env_exports,omitempty"`
	PathAdditions []string          `json:"path_additions,omitempty" yaml:"path_additions,omitempty"`
	Aliases       map[string]string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

type AppSet struct {
	Apt              []PkgItem    `json:"apt,omitempty" yaml:"apt,omitempty"`
	Dnf              []PkgItem    `json:"dnf,omitempty" yaml:"dnf,omitempty"`
	Pacman           []PacmanItem `json:"pacman,omitempty" yaml:"pacman,omitempty"`
	Zypper           []PkgItem    `json:"zypper,omitempty" yaml:"zypper,omitempty"`
	Snap             []SnapItem   `json:"snap,omitempty" yaml:"snap,omitempty"`
	Flatpak          []FlatpakItem`json:"flatpak,omitempty" yaml:"flatpak,omitempty"`
	NpmGlobal        []string     `json:"npm_global,omitempty" yaml:"npm_global,omitempty"`
	PipUser          []string     `json:"pip_user,omitempty" yaml:"pip_user,omitempty"`
	Cargo            []string     `json:"cargo,omitempty" yaml:"cargo,omitempty"`
	VSCodeExtensions []string     `json:"vscode_extensions,omitempty" yaml:"vscode_extensions,omitempty"`
}

type PkgItem struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	GUI     bool   `json:"gui,omitempty" yaml:"gui,omitempty"`
}

type PacmanItem struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Repo    string `json:"repo,omitempty" yaml:"repo,omitempty"` // e.g. "AUR"
	GUI     bool   `json:"gui,omitempty" yaml:"gui,omitempty"`
}

type SnapItem struct {
	Name    string `json:"name" yaml:"name"`
	Classic bool   `json:"classic,omitempty" yaml:"classic,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type FlatpakItem struct {
	ID      string `json:"id" yaml:"id"`
	Origin  string `json:"origin,omitempty" yaml:"origin,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type Project struct {
	Path           string    `json:"path" yaml:"path"`
	Type           string    `json:"type" yaml:"type"`
	PackageManager string    `json:"package_manager,omitempty" yaml:"package_manager,omitempty"`
	Rebuild        []string  `json:"rebuild" yaml:"rebuild"`
	Git            *GitMeta  `json:"git,omitempty" yaml:"git,omitempty"`
}

type GitMeta struct {
	Remote string `json:"remote,omitempty" yaml:"remote,omitempty"`
	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`
}

type Checksums struct {
	ManifestSHA256 string `json:"manifest_sha256,omitempty" yaml:"manifest_sha256,omitempty"`
}

// PublicMeta is the unencrypted metadata written to manifest.public.json
type PublicMeta struct {
	ID               string    `json:"id"`
	Created          time.Time `json:"created"`
	CarryoverVersion string    `json:"carryover_version"`
	SourceOS         string    `json:"source_os"`
	OSVersion        string    `json:"os_version"`
	PackageManager   string    `json:"package_manager"`
	Hostname         string    `json:"hostname"`
	TotalFolders     int       `json:"total_folders"`
	TotalApps        int       `json:"total_apps"`
	TotalProjects    int       `json:"total_projects"`
}

// NewManifest creates a new empty manifest with schema version 1
func NewManifest() *Manifest {
	return &Manifest{
		SchemaVersion: 1,
		Exclude: Exclude{
			Global: []string{
				"node_modules",
				".gradle",
				"build",
				"dist",
				"target",
				"venv",
				".venv",
				"__pycache__",
				".pytest_cache",
				".mypy_cache",
				".next",
				".nuxt",
				".turbo",
				".dart_tool",
				"Pods",
				".cache",
				".tox",
				"*.pyc",
				"*.log",
				".DS_Store",
				"Thumbs.db",
				".Trash*",
				"*.tmp",
			},
			PerProject: make(map[string][]string),
		},
		Shell: ShellConfig{
			EnvExports: make(map[string]string),
			Aliases:    make(map[string]string),
		},
	}
}

// Validate validates JSON or parsed manifest against schema
func (m *Manifest) Validate() error {
	if compiledSchema == nil {
		if schemaErr != nil {
			return fmt.Errorf("schema compilation error: %w", schemaErr)
		}
		return fmt.Errorf("schema not initialized")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for validation: %w", err)
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	if err := compiledSchema.Validate(v); err != nil {
		return fmt.Errorf("manifest schema validation failed: %w", err)
	}
	return nil
}

// SaveYAML saves the manifest to a YAML file
func (m *Manifest) SaveYAML(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest to yaml: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadYAML loads a manifest from a YAML file
func LoadYAML(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}
	return &m, nil
}

// SavePublicJSON saves the public metadata to manifest.public.json
func (m *Manifest) SavePublicJSON(path string) error {
	totalApps := len(m.Apps.Apt) + len(m.Apps.Dnf) + len(m.Apps.Pacman) +
		len(m.Apps.Zypper) + len(m.Apps.Snap) + len(m.Apps.Flatpak) +
		len(m.Apps.NpmGlobal) + len(m.Apps.PipUser) + len(m.Apps.Cargo)

	pub := PublicMeta{
		ID:               m.BackupMeta.ID,
		Created:          m.BackupMeta.Created,
		CarryoverVersion: m.BackupMeta.CarryoverVersion,
		SourceOS:         m.BackupMeta.Source.OSID,
		OSVersion:        m.BackupMeta.Source.OSVersion,
		PackageManager:   m.BackupMeta.Source.PackageManager,
		Hostname:         m.BackupMeta.Source.Hostname,
		TotalFolders:     len(m.Folders),
		TotalApps:        totalApps,
		TotalProjects:    len(m.Projects),
	}

	data, err := json.MarshalIndent(pub, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
