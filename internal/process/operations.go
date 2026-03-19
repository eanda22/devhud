package process

import (
	"fmt"
	"syscall"
)

// stops a process with SIGTERM.
func Stop(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop process: %w", err)
	}
	return nil
}
