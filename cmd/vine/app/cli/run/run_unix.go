// +build !windows

package build

import (
	"os"
	"os/exec"
	"syscall"
)

func (r *Runner) init() {
	r.cmd = exec.Command("go", r.args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
