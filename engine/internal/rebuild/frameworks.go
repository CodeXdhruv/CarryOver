package rebuild

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/carryover/carryover/engine/internal/events"
)

// ReinstallFramework installs or restores a shell framework before rc files are placed
func ReinstallFramework(homeDir, framework string, emitter *events.Emitter) error {
	if emitter != nil {
		emitter.Progress("restore_frameworks", 0, 0, 0, fmt.Sprintf("Setting up shell framework: %s", framework))
	}

	gitCloneIfMissing := func(targetPath, repoURL string) error {
		full := filepath.Join(homeDir, targetPath)
		if fi, err := os.Stat(full); err == nil && fi.IsDir() {
			return nil // Already installed
		}
		cmd := exec.Command("git", "clone", "--depth=1", repoURL, full)
		return cmd.Run()
	}

	switch framework {
	case "nvm":
		return gitCloneIfMissing(".nvm", "https://github.com/nvm-sh/nvm.git")
	case "oh-my-zsh":
		return gitCloneIfMissing(".oh-my-zsh", "https://github.com/ohmyzsh/ohmyzsh.git")
	case "oh-my-bash":
		return gitCloneIfMissing(".oh-my-bash", "https://github.com/ohmybash/oh-my-bash.git")
	case "pyenv":
		return gitCloneIfMissing(".pyenv", "https://github.com/pyenv/pyenv.git")
	case "rbenv":
		return gitCloneIfMissing(".rbenv", "https://github.com/rbenv/rbenv.git")
	case "asdf":
		return gitCloneIfMissing(".asdf", "https://github.com/asdf-vm/asdf.git")
	case "starship":
		if _, err := exec.LookPath("starship"); err == nil {
			return nil
		}
		// Install starship if absent
		cmd := exec.Command("sh", "-c", "curl -sS https://starship.rs/install.sh | sh -s -- -y --bin-dir "+filepath.Join(homeDir, ".local", "bin"))
		return cmd.Run()
	default:
		return nil
	}
}

// ReinstallAllFrameworks restores all detected shell frameworks
func ReinstallAllFrameworks(homeDir string, frameworks []string, emitter *events.Emitter) {
	for _, fw := range frameworks {
		_ = ReinstallFramework(homeDir, fw, emitter)
	}
}
