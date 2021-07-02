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
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lack-io/cli"
	"gopkg.in/fsnotify.v1"

	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
	signalutil "github.com/lack-io/vine/util/signal"
)

type Runner struct {
	wg   sync.WaitGroup
	kwg  sync.WaitGroup
	cmd  *exec.Cmd
	args []string
}

func NewRunner(args ...string) *Runner {
	return &Runner{args: args}
}

func (r *Runner) init() {
	c := exec.Command("go", r.args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	r.cmd = c
}

func (r *Runner) Run() {
	r.wg.Add(1)
	r.init()
	go func() {
		defer r.wg.Done()
		r.kwg.Wait()
		if err := r.cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}()
}

func (r *Runner) Kill() error {
	r.kwg.Add(1)
	defer r.kwg.Done()
	for {
		p := r.cmd.Process
		if p != nil {
			fmt.Printf("kill process: %d\n", p.Pid)
			return p.Kill()
		}
		time.Sleep(time.Second * 2)
	}
}

func (r *Runner) Wait() error {
	if err := r.Kill(); err != nil {
		return err
	}
	r.wg.Wait()
	return nil
}

func run(c *cli.Context) error {
	cfg, err := tool.New("vine.toml")
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("invalid vine project: %v", err)
	}

	name := c.Args().First()
	watches := c.StringSlice("watch")
	auto := c.Bool("auto-restart")
	interval := c.Int64("watch-interval")

	var mod *tool.Mod
	switch cfg.Package.Kind {
	case "cluster":
		if cfg.Mod != nil {
			for _, m := range *cfg.Mod {
				if m.Name == name {
					mod = &m
					break
				}
			}
		}
	case "single":
		if cfg.Pkg != nil && cfg.Pkg.Name == name {
			mod = cfg.Pkg
		}
	default:
		return fmt.Errorf("%s: unknown the kind of vine project", cfg.Package.Kind)
	}

	if mod == nil {
		return fmt.Errorf("not found service: " + name)
	}

	args := []string{"run", mod.Main}
	if len(c.Args().Tail()) != 0 {
		args = append(args, c.Args().Tail()...)
	}

	fmt.Printf("go %s\n", strings.Join(args, " "))
	runner := NewRunner(args...)
	if !auto {
		runner.Run()
		return nil
	}

	done := make(chan struct{})
	ech := make(chan fsnotify.Event, 1)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("start watching: %v", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case <-done:
				return
			case e, ok := <-watcher.Events:
				if !ok {
					break
				}

				switch e.Op {
				case fsnotify.Write, fsnotify.Create:
					ech <- e
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					break
				}
				fmt.Printf("watching error: %v", err)
			}
		}
	}()

	watchAdd := func(w *fsnotify.Watcher, dirs ...string) {
		for _, d := range dirs {
			_ = filepath.Walk(d, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					_ = w.Add(path)
				}

				return nil
			})
		}
	}

	switch cfg.Package.Kind {
	case "cluster":
		watchAdd(watcher, "cmd/"+name+"/", "pkg/"+name+"/", "proto/")
	case "single":
		watchAdd(watcher, "cmd/", "pkg/", "proto/")
	}

	if len(watches) > 0 {
		watchAdd(watcher, watches...)
	}

	go func() {
		t := time.Now()
		runner.Run()
		for {
			select {
			case <-done:
				goto EXIT
			case e, ok := <-ech:
				if !ok {
					goto EXIT
				}

				if e.Op == fsnotify.Create && !strings.HasSuffix(e.Name, "~") {
					stat, _ := os.Stat(e.Name)
					if stat != nil && stat.IsDir() {
						watchAdd(watcher, e.Name)
					}
					continue
				}

				now := time.Now()
				if now.Sub(t).Seconds() < float64(interval) {
					continue
				}
				t = now

				runner.Wait()
				fmt.Printf("watching change, restart go binary: go %s\n", strings.Join(args, " "))
				runner.Run()
			}
		}
	EXIT:
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signalutil.Shutdown()...)
	<-ch
	close(ch)
	return runner.Wait()
}

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "run",
			Usage: "Start a vine project",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "auto-restart",
					Usage: "auto restart project when code updating.",
					Value: true,
				},
				&cli.StringSliceFlag{
					Name:  "watch",
					Usage: "specify directory in which for watching",
				},
				&cli.Int64Flag{
					Name:    "watch-interval",
					Aliases: []string{"I"},
					Usage:   "effective interval when event triggering",
					Value:   3,
				},
			},
			Action: func(c *cli.Context) error {
				if err := run(c); err != nil {
					fmt.Println(err)
				}
				return nil
			},
		},
	}
}
