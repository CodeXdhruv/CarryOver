package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/manifest"
)

// DotfileSpec maps relative path to sensitivity
type DotfileSpec struct {
	RelPath   string
	Sensitive bool
}

// DefaultDotfileAllowlist defines standard dotfiles to look for
var DefaultDotfileAllowlist = []DotfileSpec{
	{RelPath: ".bashrc"},
	{RelPath: ".bash_profile"},
	{RelPath: ".bash_aliases"},
	{RelPath: ".profile"},
	{RelPath: ".zshrc"},
	{RelPath: ".zprofile"},
	{RelPath: ".zshenv"},
	{RelPath: ".config/fish"},
	{RelPath: ".gitconfig"},
	{RelPath: ".gitignore_global"},
	{RelPath: ".vimrc"},
	{RelPath: ".config/nvim"},
	{RelPath: ".tmux.conf"},
	{RelPath: ".ssh", Sensitive: true},
	{RelPath: ".gnupg", Sensitive: true},
	{RelPath: ".config/Code/User/settings.json"},
	{RelPath: ".config/Code/User/keybindings.json"},
	{RelPath: ".config/Code/User/snippets"},
	{RelPath: ".config/starship.toml"},
	{RelPath: ".config/alacritty"},
	{RelPath: ".config/kitty"},
	{RelPath: ".inputrc"},
	{RelPath: ".npmrc", Sensitive: true},
	{RelPath: ".aws", Sensitive: true},
	{RelPath: ".kube", Sensitive: true},
	{RelPath: ".local/share/fonts"},
	{RelPath: ".fonts"},
	{RelPath: ".themes"},
	{RelPath: ".icons"},
}

// ScanDotfiles discovers dotfiles present in user home
func ScanDotfiles(homeDir string) []manifest.Dotfile {
	var dotfiles []manifest.Dotfile

	for _, spec := range DefaultDotfileAllowlist {
		fullPath := filepath.Join(homeDir, spec.RelPath)
		fi, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		var size int64
		if fi.IsDir() {
			size = CalculateDirSize(fullPath)
		} else {
			size = fi.Size()
		}

		dotfiles = append(dotfiles, manifest.Dotfile{
			Path:      spec.RelPath,
			Sensitive: spec.Sensitive,
			SizeBytes: size,
		})
	}

	return dotfiles
}

// IsSensitivePath returns true if relative path contains known sensitive directories
func IsSensitivePath(relPath string) bool {
	clean := filepath.Clean(relPath)
	sensitivePrefixes := []string{
		".ssh",
		".gnupg",
		".npmrc",
		".aws",
		".kube",
	}
	for _, p := range sensitivePrefixes {
		if clean == p || strings.HasPrefix(clean, p+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
