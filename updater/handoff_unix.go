//go:build !windows

package updater

import (
	"os/exec"
	"syscall"
)

// configureDetached fully detaches the worker from the parent's terminal
// session via setsid and discards stdio so the parent can exit cleanly.
func configureDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
}
