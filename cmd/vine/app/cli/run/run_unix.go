//go:build !windows

package build

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func (r *Runner) init() {
	app := r.args[0]
	exec.Command("go", "build", "-o", r.tmp, app).CombinedOutput()
	args := append([]string{r.tmp}, r.args[1:]...)
	r.cmd = exec.Command("/bin/bash", "-c", strings.Join(args, " "))
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
