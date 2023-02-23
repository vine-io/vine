// MIT License
//
// Copyright (c) 2021 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package build

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vine-io/vine/cmd/vine/app/cli/util/tool"
	"github.com/vine-io/vine/util/helper"
)

func runSRV(c *cobra.Command, args []string) error {
	cfg, err := tool.New("vine.toml")
	if err != nil {
		return fmt.Errorf("invalid vine project: %v\n", err)
	}

	switch cfg.Package.Namespace {
	case "cluster":
		if cfg.Mod == nil {
			return fmt.Errorf("invalid vine project: Please create a module first.")
		}
	case "single":
		if cfg.Pkg == nil {
			return fmt.Errorf("invalid vine project: Please create a module first.")
		}
	}

	cFlags := c.PersistentFlags()
	flags, _ := cFlags.GetStringSlice("flags")
	output, _ := cFlags.GetString("output")
	name := helper.NewArgs(args).First()
	cluster := cfg.Package.Kind == "cluster"

	goPath := build.Default.GOPATH
	// attempt to split path if not windows
	if runtime.GOOS == "windows" {
		goPath = strings.Split(goPath, ";")[0]
	} else {
		goPath = strings.Split(goPath, ":")[0]
	}

	if name != "" {
		var mod *tool.Mod
		switch cfg.Package.Kind {
		case "cluster":
			for _, m := range *cfg.Mod {
				if m.Name == name {
					mod = &m
					break
				}
			}
		case "single":
			mod = cfg.Pkg
		}

		if mod == nil {
			return fmt.Errorf("module %s not found\n", name)
		}

		buildFunc(mod, output, flags, cluster)
	} else {
		switch cfg.Package.Kind {
		case "cluster":
			for _, mod := range *cfg.Mod {
				buildFunc(&mod, output, flags, cluster)
			}
		case "single":
			buildFunc(cfg.Pkg, output, flags, cluster)
		}

	}

	return nil
}

func buildFunc(mod *tool.Mod, output string, flags []string, cluster bool) {
	flags1, flags2 := []string{}, []string{}
	if len(flags) == 0 {
		flags = mod.Flags
	}

	for _, flag := range flags {
		if strings.TrimSpace(flag) == "" {
			continue
		}
		prefix := strings.Split(flag, "=")[0]
		if strings.Contains(flag, "=") && isUp(prefix) {
			flags1 = append(flags1, parseFlag(flag))
		} else {
			flags2 = append(flags2, parseFlag(flag))
		}
	}

	args := []string{"go", "build"}
	if output != "" {
		args = append(args, "-o", output)
	} else if mod.Output != "" {
		args = append(args, "-o", mod.Output)
	}
	args = append(args, flags2...)
	args = append(args, mod.Main)

	now := time.Now()

	var out []byte
	var err error
	switch runtime.GOOS {
	case "windows":
		ft := fmt.Sprintf("%s.bat", uuid.New().String())
		fmt.Printf("%s\n", strings.Join(append(flags1, args...), " "))
		_ = ioutil.WriteFile(ft, []byte(strings.Join(args, " ")), 0755)
		cmd := exec.Command("cmd", "/C", ft)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, flags1...)
		out, err = cmd.CombinedOutput()
		_ = os.Remove(ft)
	default:
		args = append(flags1, args...)
		fmt.Printf("%s\n", strings.Join(args, " "))
		cmd := exec.Command("/bin/sh", "-c", strings.Join(args, " "))
		cmd.Env = os.Environ()
		out, err = cmd.CombinedOutput()
	}
	if err != nil {
		fmt.Printf("build %s: %v\n", mod.Name, string(out))
		return
	}
	fmt.Printf("speed: %v\n", time.Now().Sub(now))
}

func parseFlag(s string) string {
	shell := func(text string, i, j int) string {
		sub := text[i+2 : j]
		parts := strings.Split(sub, " ")
		data, _ := exec.Command(parts[0], parts[1:]...).CombinedOutput()
		var out string
		switch runtime.GOOS {
		case "windows":
			out = strings.TrimSuffix(string(data), "\r\n")
		default:
			out = strings.TrimSuffix(string(data), "\n")

		}
		return text[:i] + out + text[j+1:]
	}

	c := strings.Count(s, "$")
	if c == 0 {
		return s
	}

	var out string
	for i := 0; i < c; i++ {
		a := strings.Index(s, "$")
		b := strings.Index(s, ")")
		out += shell(s[:b+1], a, b)
		s = s[b+1:]
	}

	return out
}

func isUp(text string) bool {
	b, _ := regexp.MatchString(`[A-Z_]+`, text)
	return b
}

func cmdSRV() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "service",
		Short:         "Build vine project",
		SilenceErrors: true,
		RunE:          runSRV,
	}
	cmd.PersistentFlags().StringP("flag", "L", "", "specify flags for go command.")
	cmd.PersistentFlags().StringP("output", "O", "", "specify the output path")

	return cmd
}
