package mapper

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed data/packages.yaml
var defaultPackagesYAML []byte

// Confidence levels for mapping resolution
type Confidence string

const (
	ConfidenceExact       Confidence = "exact"
	ConfidenceRepoSame    Confidence = "repo-same"
	ConfidenceTransformed Confidence = "transformed"
	ConfidenceSuggested   Confidence = "suggested"
	ConfidenceNone        Confidence = "none"
)

// Resolution is the outcome of mapping a package across managers
type Resolution struct {
	OriginalName string     `json:"original_name"`
	TargetName   string     `json:"target_name"`
	FromPM       string     `json:"from_pm"`
	ToPM         string     `json:"to_pm"`
	Confidence   Confidence `json:"confidence"`
	AutoSelect   bool       `json:"auto_select"`
	Reason       string     `json:"reason"`
}

// PackageEntry represents a package mapping entry across distributions
type PackageEntry map[string]string

// Mapper maintains mapping catalog and resolution algorithms
type Mapper struct {
	Catalog   map[string]PackageEntry
	Overrides map[string]PackageEntry
}

// NewMapper creates and initializes a package mapper with embedded curated data
func NewMapper() *Mapper {
	m := &Mapper{
		Catalog:   make(map[string]PackageEntry),
		Overrides: make(map[string]PackageEntry),
	}

	_ = yaml.Unmarshal(defaultPackagesYAML, &m.Catalog)

	// Load user overrides from ~/.config/carryover/mappings.yaml if available
	if home, err := os.UserHomeDir(); err == nil {
		userPath := filepath.Join(home, ".config", "carryover", "mappings.yaml")
		if data, err := os.ReadFile(userPath); err == nil {
			_ = yaml.Unmarshal(data, &m.Overrides)
		}
	}

	return m
}

// Resolve maps a package name from one package manager to another
func (m *Mapper) Resolve(pkgName, fromPM, toPM string) Resolution {
	fromPM = strings.ToLower(fromPM)
	toPM = strings.ToLower(toPM)

	// If source and target package managers are identical
	if fromPM == toPM {
		return Resolution{
			OriginalName: pkgName,
			TargetName:   pkgName,
			FromPM:       fromPM,
			ToPM:         toPM,
			Confidence:   ConfidenceExact,
			AutoSelect:   true,
			Reason:       "Source and target managers are identical",
		}
	}

	// 1. Check user overrides first
	if entry, ok := m.Overrides[pkgName]; ok {
		if mapped, found := entry[toPM]; found && mapped != "" {
			return Resolution{
				OriginalName: pkgName,
				TargetName:   mapped,
				FromPM:       fromPM,
				ToPM:         toPM,
				Confidence:   ConfidenceExact,
				AutoSelect:   true,
				Reason:       "Matched in user configuration overrides",
			}
		}
	}

	// 2. Check exact matches in curated catalog
	// Check by canonical key
	if entry, ok := m.Catalog[pkgName]; ok {
		if mapped, found := entry[toPM]; found && mapped != "" {
			return Resolution{
				OriginalName: pkgName,
				TargetName:   mapped,
				FromPM:       fromPM,
				ToPM:         toPM,
				Confidence:   ConfidenceExact,
				AutoSelect:   true,
				Reason:       "Exact match in curated catalog",
			}
		}
	}

	// Check by value in fromPM
	for _, entry := range m.Catalog {
		if entry[fromPM] == pkgName {
			if mapped, found := entry[toPM]; found && mapped != "" {
				return Resolution{
					OriginalName: pkgName,
					TargetName:   mapped,
					FromPM:       fromPM,
					ToPM:         toPM,
					Confidence:   ConfidenceExact,
					AutoSelect:   true,
					Reason:       "Matched via source package in curated catalog",
				}
			}
		}
	}

	// 3. Known suffix and prefix transformations
	transformed := transformPackageName(pkgName, fromPM, toPM)
	if transformed != "" && transformed != pkgName {
		return Resolution{
			OriginalName: pkgName,
			TargetName:   transformed,
			FromPM:       fromPM,
			ToPM:         toPM,
			Confidence:   ConfidenceTransformed,
			AutoSelect:   true,
			Reason:       "Heuristic naming convention transform",
		}
	}

	// 4. Same name in target repo fallback
	return Resolution{
		OriginalName: pkgName,
		TargetName:   pkgName,
		FromPM:       fromPM,
		ToPM:         toPM,
		Confidence:   ConfidenceRepoSame,
		AutoSelect:   true,
		Reason:       "Direct name assumption in target distribution",
	}
}

// transformPackageName applies standard packaging heuristics across distros
func transformPackageName(name, fromPM, toPM string) string {
	// Dev headers: -dev (Debian/Ubuntu) <-> -devel (Fedora/openSUSE)
	if (fromPM == "apt") && (toPM == "dnf" || toPM == "zypper") {
		if strings.HasSuffix(name, "-dev") {
			return strings.TrimSuffix(name, "-dev") + "-devel"
		}
	}
	if (fromPM == "dnf" || fromPM == "zypper") && (toPM == "apt") {
		if strings.HasSuffix(name, "-devel") {
			return strings.TrimSuffix(name, "-devel") + "-dev"
		}
	}

	// Python naming: python3-foo (Debian/Fedora) <-> python-foo (Arch)
	if (fromPM == "apt" || fromPM == "dnf" || fromPM == "zypper") && toPM == "pacman" {
		if strings.HasPrefix(name, "python3-") {
			return "python-" + strings.TrimPrefix(name, "python3-")
		}
	}
	if fromPM == "pacman" && (toPM == "apt" || toPM == "dnf" || toPM == "zypper") {
		if strings.HasPrefix(name, "python-") {
			return "python3-" + strings.TrimPrefix(name, "python-")
		}
	}

	return ""
}
