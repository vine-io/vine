// MIT License
//
// Copyright (c) 2020 Lack
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

package bot

import (
	"strings"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/lib/agent/command"
	clic "github.com/lack-io/vine/util/command/cli"
)

// Echo returns the same message
func Echo(ctx *cli.Context) command.Command {
	usage := "echo [text]"
	desc := "Returns the [text]"

	return command.NewCommand("echo", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("echo what?"), nil
		}
		return []byte(strings.Join(args[1:], " ")), nil
	})
}

// Hello returns a greeting
func Hello(ctx *cli.Context) command.Command {
	usage := "hello"
	desc := "Returns a greeting"

	return command.NewCommand("hello", usage, desc, func(args ...string) ([]byte, error) {
		return []byte("hey what's up?"), nil
	})
}

// Ping returns pong
func Ping(ctx *cli.Context) command.Command {
	usage := "ping"
	desc := "Returns pong"

	return command.NewCommand("ping", usage, desc, func(args ...string) ([]byte, error) {
		return []byte("pong"), nil
	})
}

// Get service returns a service
func Get(ctx *cli.Context) command.Command {
	usage := "get service [name]"
	desc := "Returns a registered service"

	return command.NewCommand("get", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("get what?"), nil
		}
		switch args[1] {
		case "service":
			if len(args) < 3 {
				return []byte("require service name"), nil
			}
			rsp, err := clic.GetService(ctx, args[2:])
			if err != nil {
				return nil, err
			}
			return rsp, nil
		default:
			return []byte("unknown command...\nsupported commands: \nget service [name]"), nil
		}
	})
}

// Health returns the health of a service
func Health(ctx *cli.Context) command.Command {
	usage := "health [service]"
	desc := "Returns health of a service"

	return command.NewCommand("health", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("health of what?"), nil
		}
		rsp, err := clic.QueryHealth(ctx, args[1:])
		if err != nil {
			return nil, err
		}
		return rsp, nil
	})
}

// List returns a list of services
func List(ctx *cli.Context) command.Command {
	usage := "list services"
	desc := "Returns a list of registered services"

	return command.NewCommand("list", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("list what?"), nil
		}
		switch args[1] {
		case "services":
			rsp, err := clic.ListServices(ctx)
			if err != nil {
				return nil, err
			}
			return rsp, nil
		default:
			return []byte("unknown command...\nsupported commands: \nlist services"), nil
		}
	})
}

// Call returns a service call
func Call(ctx *cli.Context) command.Command {
	usage := "call [service] [endpoint] [request]"
	desc := "Returns the response for a service call"

	return command.NewCommand("call", usage, desc, func(args ...string) ([]byte, error) {
		var cargs []string

		for _, arg := range args {
			if len(strings.TrimSpace(arg)) == 0 {
				continue
			}
			cargs = append(cargs, arg)
		}

		if len(cargs) < 2 {
			return []byte("call what?"), nil
		}

		rsp, err := clic.CallService(ctx, cargs[1:])
		if err != nil {
			return nil, err
		}
		return rsp, nil
	})
}

// Register registers a service
func Register(ctx *cli.Context) command.Command {
	usage := "register service [definition]"
	desc := "Registers a service"

	return command.NewCommand("register", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("register what?"), nil
		}
		switch args[1] {
		case "service":
			if len(args) < 3 {
				return []byte("require service definition"), nil
			}
			rsp, err := clic.RegisterService(ctx, args[2:])
			if err != nil {
				return nil, err
			}
			return rsp, nil
		default:
			return []byte("unknown command...\nsupported commands: \nregister service [definition]"), nil
		}
	})
}

// Deregister registers a service
func Deregister(ctx *cli.Context) command.Command {
	usage := "deregister service [definition]"
	desc := "Deregisters a service"

	return command.NewCommand("deregister", usage, desc, func(args ...string) ([]byte, error) {
		if len(args) < 2 {
			return []byte("deregister what?"), nil
		}
		switch args[1] {
		case "service":
			if len(args) < 3 {
				return []byte("require service definition"), nil
			}
			rsp, err := clic.DeregisterService(ctx, args[2:])
			if err != nil {
				return nil, err
			}
			return rsp, nil
		default:
			return []byte("unknown command...\nsupported commands: \nderegister service [definition]"), nil
		}
	})
}

// Laws of robotics
func ThreeLaws(ctx *cli.Context) command.Command {
	usage := "the three laws"
	desc := "Returns the three laws of robotics"

	return command.NewCommand("the three laws", usage, desc, func(args ...string) ([]byte, error) {
		laws := []string{
			"1. A robot may not injure a human being or, through inaction, allow a human being to come to harm.",
			"2. A robot must obey the orders given it by human beings except where such orders would conflict with the First Law.",
			"3. A robot must protect its own existence as long as such protection does not conflict with the First or Second Laws.",
		}
		return []byte("\n" + strings.Join(laws, "\n")), nil
	})
}

// Time returns the time
func Time(ctx *cli.Context) command.Command {
	usage := "time"
	desc := "Returns the server time"

	return command.NewCommand("time", usage, desc, func(args ...string) ([]byte, error) {
		t := time.Now().Format(time.RFC1123)
		return []byte("Server time is: " + t), nil
	})
}
