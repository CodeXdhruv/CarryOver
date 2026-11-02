package scanner

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/carryover/carryover/engine/internal/manifest"
)

var (
	exportRegex = regexp.MustCompile(`^export\s+([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)
	aliasRegex  = regexp.MustCompile(`^alias\s+([A-Za-z0-9_-]+)=(.*)$`)
)

// ParseShellRC statically parses an rc file (like .bashrc or .zshrc) for exports, aliases, and paths
func ParseShellRC(r io.Reader) (env map[string]string, aliases map[string]string, paths []string) {
	env = make(map[string]string)
	aliases = make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check exports
		if m := exportRegex.FindStringSubmatch(line); len(m) == 3 {
			k := m[1]
			v := strings.Trim(strings.TrimSpace(m[2]), `"'`)
			if k == "PATH" {
				paths = append(paths, v)
			} else {
				env[k] = v
			}
			continue
		}

		// Check aliases
		if m := aliasRegex.FindStringSubmatch(line); len(m) == 3 {
			k := m[1]
			v := strings.Trim(strings.TrimSpace(m[2]), `"'`)
			aliases[k] = v
			continue
		}
	}

	return env, aliases, paths
}

// DetectShellFrameworks checks filesystem markers and environment variables for known shell frameworks
func DetectShellFrameworks(homeDir string) []string {
	var frameworks []string

	checks := []struct {
		name string
		path string
		env  string
	}{
		{name: "oh-my-zsh", path: ".oh-my-zsh", env: "ZSH"},
		{name: "oh-my-bash", path: ".oh-my-bash"},
		{name: "starship", path: ".config/starship.toml"},
		{name: "nvm", path: ".nvm", env: "NVM_DIR"},
		{name: "pyenv", path: ".pyenv", env: "PYENV_ROOT"},
		{name: "rbenv", path: ".rbenv"},
		{name: "sdkman", path: ".sdkman", env: "SDKMAN_DIR"},
		{name: "asdf", path: ".asdf", env: "ASDF_DATA_DIR"},
		{name: "conda", path: ".conda", env: "CONDA_PREFIX"},
	}

	for _, c := range checks {
		found := false
		if c.env != "" && os.Getenv(c.env) != "" {
			found = true
		} else if c.path != "" {
			full := filepath.Join(homeDir, c.path)
			if _, err := os.Stat(full); err == nil {
				found = true
			}
		}
		if found {
			frameworks = append(frameworks, c.name)
		}
	}

	return frameworks
}

// ScanShellConfig gathers shell configuration, frameworks, environment exports, and aliases
func ScanShellConfig(homeDir, defaultShell string) manifest.ShellConfig {
	cfg := manifest.ShellConfig{
		Default:    defaultShell,
		EnvExports: make(map[string]string),
		Aliases:    make(map[string]string),
	}

	rcFiles := []string{".bashrc", ".zshrc", ".profile"}
	for _, rc := range rcFiles {
		fullPath := filepath.Join(homeDir, rc)
		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}
		env, aliases, paths := ParseShellRC(f)
		_ = f.Close()

		for k, v := range env {
			cfg.EnvExports[k] = v
		}
		for k, v := range aliases {
			cfg.Aliases[k] = v
		}
		cfg.PathAdditions = append(cfg.PathAdditions, paths...)
	}

	cfg.Frameworks = DetectShellFrameworks(homeDir)
	return cfg
}
