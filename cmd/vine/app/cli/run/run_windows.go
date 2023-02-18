//go:build windows

package build

import (
	"os"
	"os/exec"
	"strings"
)

func (r *Runner) init() {
	app := r.args[0]
	exec.Command("go", "build", "-o", r.tmp, app).CombinedOutput()
	args := append([]string{r.tmp}, r.args[1:]...)
	r.cmd = exec.Command("cmd", "/C", strings.Join(args, " "))
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
}
