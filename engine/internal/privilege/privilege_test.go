package privilege

import (
	"strings"
	"testing"
)

func TestBuildElevatedCommand(t *testing.T) {
	cmd := []string{"apt-get", "install", "-y", "git"}
	elevated, preview := BuildElevatedCommand(cmd)

	if preview != "apt-get install -y git" {
		t.Errorf("expected clean preview string, got: %s", preview)
	}
	if len(elevated) < len(cmd) {
		t.Errorf("expected elevated command length >= original command length")
	}

	// Last arguments should match original command
	tail := elevated[len(elevated)-len(cmd):]
	if strings.Join(tail, " ") != strings.Join(cmd, " ") {
		t.Errorf("expected command tail to match original: %v vs %v", tail, cmd)
	}
}
