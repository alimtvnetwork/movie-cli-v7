//go:build windows

package updater

import (
	"os/exec"
	"syscall"
)

// Windows process creation flags. Defined locally to avoid an external
// dependency on golang.org/x/sys/windows.
const (
	createNewConsole      = 0x00000010
	createNewProcessGroup = 0x00000200
)

// configureDetached makes the worker run in its own visible console window
// and its own process group. The new console is what keeps the worker's
// progress output visible to the user even though the parent has exited.
func configureDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNewConsole | createNewProcessGroup,
	}
}
