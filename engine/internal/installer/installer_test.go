package installer

import (
	"strings"
	"testing"
)

func TestBuildNativeCommand(t *testing.T) {
	pkgs := []string{"git", "curl", "neovim"}

	// apt
	aptCmd := BuildNativeCommand("apt", pkgs)
	if strings.Join(aptCmd[:4], " ") != "apt-get install -y --no-install-recommends" {
		t.Errorf("unexpected apt command prefix: %v", aptCmd)
	}

	// dnf
	dnfCmd := BuildNativeCommand("dnf", pkgs)
	if strings.Join(dnfCmd[:3], " ") != "dnf install -y" {
		t.Errorf("unexpected dnf command prefix: %v", dnfCmd)
	}

	// pacman
	pacmanCmd := BuildNativeCommand("pacman", pkgs)
	if strings.Join(pacmanCmd[:4], " ") != "pacman -S --needed --noconfirm" {
		t.Errorf("unexpected pacman command prefix: %v", pacmanCmd)
	}

	// zypper
	zypperCmd := BuildNativeCommand("zypper", pkgs)
	if strings.Join(zypperCmd[:3], " ") != "zypper install -y" {
		t.Errorf("unexpected zypper command prefix: %v", zypperCmd)
	}
}
