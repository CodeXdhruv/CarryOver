package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/carryover/carryover/engine/internal/manifest"
)

// DefaultFolderCandidates defines the standard relative paths under $HOME to scan
var DefaultFolderCandidates = []string{
	"Documents",
	"Pictures",
	"Desktop",
	"Music",
	"Videos",
	"Downloads",
	"Projects",
	"dev",
	"code",
	"workspace",
	"src",
}

// GlobalSkipDirectories lists directory basenames to exclude during sizing and backup
var GlobalSkipDirectories = map[string]bool{
	"node_modules":  true,
	".gradle":       true,
	"build":         true,
	"dist":          true,
	"target":        true,
	"venv":          true,
	".venv":         true,
	"__pycache__":   true,
	".pytest_cache": true,
	".mypy_cache":   true,
	".next":         true,
	".nuxt":         true,
	".turbo":        true,
	".dart_tool":    true,
	"Pods":          true,
	".cache":        true,
	".tox":          true,
	".Trash":        true,
	".Trash-1000":   true,
}

// CalculateDirSize returns total bytes in dir, skipping rebuildable directories
func CalculateDirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (GlobalSkipDirectories[name] || strings.HasPrefix(name, ".")) {
				// Skip rebuildable or heavy hidden cache directories
				if GlobalSkipDirectories[name] || name == ".cache" || name == ".git" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// ScanFolders scans user home for candidate directories and estimates size
func ScanFolders(homeDir string) []manifest.Folder {
	var folders []manifest.Folder
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, rel := range DefaultFolderCandidates {
		fullPath := filepath.Join(homeDir, rel)
		fi, err := os.Stat(fullPath)
		if err != nil || !fi.IsDir() {
			continue
		}

		wg.Add(1)
		go func(relative, full string) {
			defer wg.Done()
			size := CalculateDirSize(full)
			mu.Lock()
			folders = append(folders, manifest.Folder{
				Path:      relative,
				SizeBytes: size,
			})
			mu.Unlock()
		}(rel, fullPath)
	}

	wg.Wait()
	return folders
}
