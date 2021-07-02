// +build windows

package build

import (
	"os"
	"os/exec"
)

func (r *Runner) init() {
	r.cmd = exec.Command("go", r.args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
}
