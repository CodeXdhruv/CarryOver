package privilege

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/carryover/carryover/engine/internal/events"
)

// IsRoot checks if current effective user is root
func IsRoot() bool {
	return os.Geteuid() == 0
}

// BuildElevatedCommand wraps a command with pkexec or sudo if running unprivileged
func BuildElevatedCommand(cmd []string) ([]string, string) {
	fullPreview := strings.Join(cmd, " ")
	if IsRoot() {
		return cmd, fullPreview
	}

	// Try pkexec first
	if pkexecPath, err := exec.LookPath("pkexec"); err == nil {
		elevated := append([]string{pkexecPath}, cmd...)
		return elevated, fullPreview
	}

	// Fallback to sudo
	if sudoPath, err := exec.LookPath("sudo"); err == nil {
		elevated := append([]string{sudoPath}, cmd...)
		return elevated, fullPreview
	}

	// Return as-is if no elevation tool available
	return cmd, fullPreview
}

// ExecutePrivileged batches and runs an elevated command with user preview
func ExecutePrivileged(cmd []string, emitter *events.Emitter, reason string) error {
	elevated, preview := BuildElevatedCommand(cmd)

	if emitter != nil {
		emitter.NeedPrivilege(reason, preview)
	}

	c := exec.Command(elevated[0], elevated[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("privileged command failed (%s): %w", preview, err)
	}
	return nil
}
