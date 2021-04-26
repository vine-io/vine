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
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/lack-io/cli"
	"gopkg.in/fsnotify.v1"

	"github.com/lack-io/vine/cmd/vine/app/cli/util/tool"
	signalutil "github.com/lack-io/vine/util/signal"
)

func run(c *cli.Context) error {
	cfg, err := tool.New("vine.toml")
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("invalid vine project: %v", err)
	}

	name := c.Args().First()
	watches := c.StringSlice("watch")
	auto := c.Bool("auto-restart")

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
	if !auto {
		return goRun(context.Background(), args)
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
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := goRun(ctx, args); err != nil {
				fmt.Println(err)
			}
		}()
		for {
			select {
			case <-done:
				goto EXIT
			case e, ok := <-ech:
				if !ok {
					goto EXIT
				}
				cancel()
				time.Sleep(time.Millisecond * 500)
				ctx, cancel = context.WithCancel(context.Background())

				if e.Op == fsnotify.Create && !strings.HasSuffix(e.Name, "~") {
					stat, _ := os.Stat(e.Name)
					if stat != nil && stat.IsDir() {
						watchAdd(watcher, e.Name)
					}
					continue
				}

				fmt.Printf("watching change, restart go binary: go %s\n", strings.Join(args, " "))
				go func() {
					if err := goRun(ctx, args); err != nil && err != context.Canceled && !strings.HasPrefix(err.Error(), "signal") {
						fmt.Println(err)
					}
				}()
			}
		}
	EXIT:
		cancel()
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signalutil.Shutdown()...)
	<-ch
	close(ch)
	return nil
}

func goRun(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "run",
			Usage: "Start a vine project",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "auto-restart",
					Usage: "auto restart project when code updated.",
					Value: true,
				},
				&cli.StringSliceFlag{
					Name:  "watch",
					Usage: "specify directory in which for watching",
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
