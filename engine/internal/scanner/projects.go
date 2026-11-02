package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carryover/carryover/engine/internal/manifest"
)

// ProjectRecipe defines how a detected project type is classified and rebuilt
type ProjectRecipe struct {
	Type           string
	PackageManager string
	Rebuild        []string
	Excludes       []string
}

// DetectProjectInDir inspects marker files inside a directory to derive recipe
func DetectProjectInDir(dir string) *ProjectRecipe {
	hasFile := func(name string) bool {
		matches, err := filepath.Glob(filepath.Join(dir, name))
		return err == nil && len(matches) > 0
	}

	// Node.js
	if hasFile("package.json") {
		excludes := []string{"node_modules", ".next", ".nuxt", "dist", ".turbo", ".cache"}
		if hasFile("package-lock.json") {
			return &ProjectRecipe{
				Type:           "node",
				PackageManager: "npm",
				Rebuild:        []string{"npm ci"},
				Excludes:       excludes,
			}
		}
		if hasFile("yarn.lock") {
			return &ProjectRecipe{
				Type:           "node",
				PackageManager: "yarn",
				Rebuild:        []string{"yarn install --frozen-lockfile"},
				Excludes:       excludes,
			}
		}
		if hasFile("pnpm-lock.yaml") {
			return &ProjectRecipe{
				Type:           "node",
				PackageManager: "pnpm",
				Rebuild:        []string{"pnpm install --frozen-lockfile"},
				Excludes:       excludes,
			}
		}
		return &ProjectRecipe{
			Type:           "node",
			PackageManager: "npm",
			Rebuild:        []string{"npm install"},
			Excludes:       excludes,
		}
	}

	// Flutter / Dart
	if hasFile("pubspec.yaml") {
		return &ProjectRecipe{
			Type:           "flutter",
			PackageManager: "flutter",
			Rebuild:        []string{"flutter pub get"},
			Excludes:       []string{".dart_tool", "build", ".pub-cache"},
		}
	}

	// Rust
	if hasFile("Cargo.toml") {
		return &ProjectRecipe{
			Type:           "rust",
			PackageManager: "cargo",
			Rebuild:        []string{"cargo fetch"},
			Excludes:       []string{"target"},
		}
	}

	// Go
	if hasFile("go.mod") {
		return &ProjectRecipe{
			Type:           "go",
			PackageManager: "go",
			Rebuild:        []string{"go mod download"},
			Excludes:       []string{"vendor"},
		}
	}

	// Python
	if hasFile("pyproject.toml") {
		return &ProjectRecipe{
			Type:           "python",
			PackageManager: "poetry",
			Rebuild:        []string{"poetry install"},
			Excludes:       []string{".venv", "venv", "__pycache__"},
		}
	}
	if hasFile("Pipfile") {
		return &ProjectRecipe{
			Type:           "python",
			PackageManager: "pipenv",
			Rebuild:        []string{"pipenv install"},
			Excludes:       []string{".venv", "venv", "__pycache__"},
		}
	}
	if hasFile("requirements.txt") {
		return &ProjectRecipe{
			Type:           "python",
			PackageManager: "pip",
			Rebuild:        []string{"python -m venv .venv && .venv/bin/pip install -r requirements.txt"},
			Excludes:       []string{"venv", ".venv", "env", "__pycache__", ".mypy_cache", ".pytest_cache"},
		}
	}

	// Java / Gradle / Maven
	if hasFile("build.gradle*") {
		return &ProjectRecipe{
			Type:           "gradle",
			PackageManager: "gradle",
			Rebuild:        []string{"./gradlew --refresh-dependencies"},
			Excludes:       []string{".gradle", "build", "app/build", ".idea"},
		}
	}
	if hasFile("pom.xml") {
		return &ProjectRecipe{
			Type:           "maven",
			PackageManager: "maven",
			Rebuild:        []string{"mvn -q dependency:resolve"},
			Excludes:       []string{"target"},
		}
	}

	// PHP Composer
	if hasFile("composer.json") {
		return &ProjectRecipe{
			Type:           "php",
			PackageManager: "composer",
			Rebuild:        []string{"composer install"},
			Excludes:       []string{"vendor"},
		}
	}

	// Ruby Bundler
	if hasFile("Gemfile") {
		return &ProjectRecipe{
			Type:           "ruby",
			PackageManager: "bundler",
			Rebuild:        []string{"bundle install"},
			Excludes:       []string{"vendor/bundle", ".bundle"},
		}
	}

	// .NET C#
	if hasFile("*.csproj") || hasFile("*.sln") {
		return &ProjectRecipe{
			Type:           "dotnet",
			PackageManager: "dotnet",
			Rebuild:        []string{"dotnet restore"},
			Excludes:       []string{"bin", "obj"},
		}
	}

	// Docker Compose
	if hasFile("docker-compose.yml") || hasFile("docker-compose.yaml") {
		return &ProjectRecipe{
			Type:           "docker",
			PackageManager: "docker-compose",
			Rebuild:        []string{"docker compose pull"},
			Excludes:       nil,
		}
	}

	return nil
}

// GetGitMeta inspects git repo inside dir if present
func GetGitMeta(dir string) *manifest.GitMeta {
	gitDir := filepath.Join(dir, ".git")
	if fi, err := os.Stat(gitDir); err != nil || (!fi.IsDir() && !fi.Mode().IsRegular()) {
		return nil
	}

	meta := &manifest.GitMeta{}
	cmdRemote := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url")
	if out, err := cmdRemote.Output(); err == nil {
		meta.Remote = strings.TrimSpace(string(out))
	}

	cmdBranch := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	if out, err := cmdBranch.Output(); err == nil {
		meta.Branch = strings.TrimSpace(string(out))
	}

	return meta
}

// ScanProjects walks project roots up to maxDepth and discovers projects
func ScanProjects(homeDir string, roots []string, maxDepth int) ([]manifest.Project, map[string][]string) {
	var projects []manifest.Project
	perProjectExcludes := make(map[string][]string)

	visited := make(map[string]bool)

	for _, root := range roots {
		fullRoot := filepath.Join(homeDir, root)
		if _, err := os.Stat(fullRoot); err != nil {
			continue
		}

		_ = filepath.WalkDir(fullRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}

			// Don't descend into skipped cache or build folders
			base := d.Name()
			if path != fullRoot && (GlobalSkipDirectories[base] || base == ".git") {
				return filepath.SkipDir
			}

			// Calculate depth relative to fullRoot
			relFromRoot, _ := filepath.Rel(fullRoot, path)
			depth := len(strings.Split(relFromRoot, string(filepath.Separator)))
			if relFromRoot == "." {
				depth = 0
			}
			if depth > maxDepth {
				return filepath.SkipDir
			}

			if visited[path] {
				return nil
			}

			recipe := DetectProjectInDir(path)
			if recipe != nil {
				visited[path] = true
				relHome, err := filepath.Rel(homeDir, path)
				if err != nil {
					relHome = path
				}

				proj := manifest.Project{
					Path:           relHome,
					Type:           recipe.Type,
					PackageManager: recipe.PackageManager,
					Rebuild:        recipe.Rebuild,
					Git:            GetGitMeta(path),
				}
				projects = append(projects, proj)
				if len(recipe.Excludes) > 0 {
					perProjectExcludes[relHome] = recipe.Excludes
				}
				// Don't walk nested submodules/projects automatically
				return filepath.SkipDir
			}

			return nil
		})
	}

	return projects, perProjectExcludes
}
