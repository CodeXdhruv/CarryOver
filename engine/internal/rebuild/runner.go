package rebuild

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/carryover/carryover/engine/internal/events"
)

// RebuildOptions configures a rebuild execution
type RebuildOptions struct {
	ProjectPath   string
	ProjectType   string
	Command       string
	Timeout       time.Duration
	AllowCustom   bool
	Emitter       *events.Emitter
}

// RunRebuild executes an allowlisted rebuild command in the project directory
func RunRebuild(opts RebuildOptions) error {
	if opts.Timeout == 0 {
		opts.Timeout = 15 * time.Minute
	}

	// Security validation: enforce allowlist unless explicitly overridden
	if !opts.AllowCustom {
		if !IsCommandAllowed(opts.ProjectType, opts.Command) {
			return fmt.Errorf("rebuild command '%s' is not in the security allowlist for project type '%s'", opts.Command, opts.ProjectType)
		}
	}

	if opts.Emitter != nil {
		opts.Emitter.Progress("rebuild", 0, 0, 0, fmt.Sprintf("Rebuilding %s: %s", opts.ProjectPath, opts.Command))
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", opts.Command)
	cmd.Dir = opts.ProjectPath

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("rebuild command timed out after %v: %s", opts.Timeout, opts.Command)
		}
		return fmt.Errorf("rebuild command failed: %s (%w)", string(out), err)
	}

	return nil
}
