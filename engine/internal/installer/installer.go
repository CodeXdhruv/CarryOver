package installer

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/carryover/carryover/engine/internal/events"
	"github.com/carryover/carryover/engine/internal/privilege"
)

// InstallItem represents an application or package to install
type InstallItem struct {
	Manager string   `json:"manager"` // apt, dnf, pacman, zypper, flatpak, snap, pip, npm, cargo, vscode
	Name    string   `json:"name"`
	Classic bool     `json:"classic,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// BuildNativeCommand constructs batched native package manager command
func BuildNativeCommand(pm string, packages []string) []string {
	if len(packages) == 0 {
		return nil
	}

	switch pm {
	case "apt":
		return append([]string{"apt-get", "install", "-y", "--no-install-recommends"}, packages...)
	case "dnf":
		return append([]string{"dnf", "install", "-y"}, packages...)
	case "pacman":
		return append([]string{"pacman", "-S", "--needed", "--noconfirm"}, packages...)
	case "zypper":
		return append([]string{"zypper", "install", "-y"}, packages...)
	default:
		return nil
	}
}

// InstallNativePackages installs a batch of system packages using privilege escalation
func InstallNativePackages(pm string, packages []string, emitter *events.Emitter) error {
	if len(packages) == 0 {
		return nil
	}
	cmd := BuildNativeCommand(pm, packages)
	if cmd == nil {
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	reason := fmt.Sprintf("Install %d packages via %s (%s)", len(packages), pm, strings.Join(packages, ", "))
	return privilege.ExecutePrivileged(cmd, emitter, reason)
}

// InstallFlatpakPackages installs Flatpak applications
func InstallFlatpakPackages(appIDs []string, emitter *events.Emitter) error {
	for _, app := range appIDs {
		if emitter != nil {
			emitter.Progress("install_apps", 0, 0, 0, fmt.Sprintf("Installing Flatpak: %s", app))
		}
		cmd := exec.Command("flatpak", "install", "-y", "flathub", app)
		if out, err := cmd.CombinedOutput(); err != nil {
			if emitter != nil {
				emitter.Warning("FLATPAK_INSTALL_WARN", fmt.Sprintf("Flatpak %s failed: %s", app, string(out)), app)
			}
		}
	}
	return nil
}

// InstallSnapPackages installs snap packages
func InstallSnapPackages(items []InstallItem, emitter *events.Emitter) error {
	for _, item := range items {
		args := []string{"snap", "install", item.Name}
		if item.Classic {
			args = append(args, "--classic")
		}
		reason := fmt.Sprintf("Install snap package %s", item.Name)
		_ = privilege.ExecutePrivileged(args, emitter, reason)
	}
	return nil
}

// InstallNpmGlobalPackages installs global npm modules
func InstallNpmGlobalPackages(pkgs []string, emitter *events.Emitter) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"npm", "install", "-g"}, pkgs...)
	reason := fmt.Sprintf("Install global npm packages: %s", strings.Join(pkgs, ", "))
	return privilege.ExecutePrivileged(args, emitter, reason)
}

// InstallPipPackages installs user-level pip packages
func InstallPipPackages(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"install", "--user"}, pkgs...)
	cmd := exec.Command("pip", args...)
	return cmd.Run()
}

// InstallCargoPackages installs cargo crates
func InstallCargoPackages(pkgs []string) error {
	for _, p := range pkgs {
		cmd := exec.Command("cargo", "install", p)
		_ = cmd.Run()
	}
	return nil
}

// InstallVSCodeExtensions installs VS Code extensions
func InstallVSCodeExtensions(exts []string) error {
	for _, e := range exts {
		cmd := exec.Command("code", "--install-extension", e)
		_ = cmd.Run()
	}
	return nil
}
