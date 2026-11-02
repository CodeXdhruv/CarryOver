package rebuild

import (
	"strings"
)

// AllowedCommandsByType defines the strict security allowlist for rebuild commands
var AllowedCommandsByType = map[string][]string{
	"node": {
		"npm ci",
		"npm install",
		"yarn install --frozen-lockfile",
		"yarn install",
		"pnpm install --frozen-lockfile",
		"pnpm install",
	},
	"flutter": {
		"flutter pub get",
		"dart pub get",
	},
	"rust": {
		"cargo fetch",
		"cargo build",
	},
	"go": {
		"go mod download",
	},
	"python": {
		"poetry install",
		"pipenv install",
		"python -m venv .venv && .venv/bin/pip install -r requirements.txt",
		"python3 -m venv .venv && .venv/bin/pip install -r requirements.txt",
	},
	"gradle": {
		"./gradlew --refresh-dependencies",
		"./gradlew build",
	},
	"maven": {
		"mvn -q dependency:resolve",
	},
	"php": {
		"composer install",
	},
	"ruby": {
		"bundle install",
	},
	"dotnet": {
		"dotnet restore",
	},
	"docker": {
		"docker compose pull",
	},
}

// IsCommandAllowed checks if a command string is permit-listed for the project type
func IsCommandAllowed(projType string, cmdStr string) bool {
	cmdStr = strings.TrimSpace(cmdStr)
	allowed, ok := AllowedCommandsByType[strings.ToLower(projType)]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == cmdStr {
			return true
		}
	}
	return false
}

// DeriveSafeCommands returns allowlisted commands for a project type and manager
func DeriveSafeCommands(projType, pkgManager string) []string {
	projType = strings.ToLower(projType)
	pkgManager = strings.ToLower(pkgManager)

	switch projType {
	case "node":
		switch pkgManager {
		case "yarn":
			return []string{"yarn install --frozen-lockfile"}
		case "pnpm":
			return []string{"pnpm install --frozen-lockfile"}
		default:
			return []string{"npm ci"}
		}
	case "flutter":
		return []string{"flutter pub get"}
	case "rust":
		return []string{"cargo fetch"}
	case "go":
		return []string{"go mod download"}
	case "python":
		switch pkgManager {
		case "poetry":
			return []string{"poetry install"}
		case "pipenv":
			return []string{"pipenv install"}
		default:
			return []string{"python -m venv .venv && .venv/bin/pip install -r requirements.txt"}
		}
	case "gradle":
		return []string{"./gradlew --refresh-dependencies"}
	case "maven":
		return []string{"mvn -q dependency:resolve"}
	case "php":
		return []string{"composer install"}
	case "ruby":
		return []string{"bundle install"}
	case "dotnet":
		return []string{"dotnet restore"}
	case "docker":
		return []string{"docker compose pull"}
	default:
		return nil
	}
}
