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

package cli

import (
	b "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lack-io/cli"
	"github.com/lack-io/vine/cmd/vine/app/cli/util"
	"github.com/lack-io/vine/core/client"
	"github.com/lack-io/vine/core/codec/bytes"
	"github.com/lack-io/vine/lib/cmd"
	clic "github.com/lack-io/vine/util/command/cli"
	"github.com/lack-io/vine/util/file"
)

type exec func(*cli.Context, []string) ([]byte, error)

func Print(e exec) func(*cli.Context) error {
	return func(c *cli.Context) error {
		rsp, err := e(c, c.Args().Slice())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(rsp) > 0 {
			fmt.Printf("%s\n", string(rsp))
		}
		return nil
	}
}

func list(c *cli.Context, args []string) ([]byte, error) {
	// no args
	if len(args) == 0 {
		return clic.ListServices(c)
	}

	// check first arg
	switch args[0] {
	case "services":
		return clic.ListServices(c)
	case "nodes":
		return clic.NetworkNodes(c)
	case "routes":
		return clic.NetworkRoutes(c)
	}

	return nil, errors.New("unknown command")
}

func networkConnect(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkConnect(c, args)
}

func networkConnections(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkConnections(c)
}

func networkGraph(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkGraph(c)
}

func networkServices(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkServices(c)
}

func netNodes(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkNodes(c)
}

func netRoutes(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkRoutes(c)
}

func netDNSAdvertise(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkDNSAdvertise(c)
}

func netDNSRemove(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkDNSRemove(c)
}

func netDNSResolve(c *cli.Context, args []string) ([]byte, error) {
	return clic.NetworkDNSResolve(c)
}

func listServices(c *cli.Context, args []string) ([]byte, error) {
	return clic.ListServices(c)
}

func registerService(c *cli.Context, args []string) ([]byte, error) {
	return clic.RegisterService(c, args)
}

func deregisterService(c *cli.Context, args []string) ([]byte, error) {
	return clic.DeregisterService(c, args)
}

func getService(c *cli.Context, args []string) ([]byte, error) {
	return clic.GetService(c, args)
}

func callService(c *cli.Context, args []string) ([]byte, error) {
	return clic.CallService(c, args)
}

func getEnv(c *cli.Context, args []string) ([]byte, error) {
	env := util.GetEnv(c)
	return []byte(env.Name), nil
}

func setEnv(c *cli.Context, args []string) ([]byte, error) {
	util.SetEnv(args[0])
	return nil, nil
}

func listEnvs(c *cli.Context, args []string) ([]byte, error) {
	envs := util.GetEnvs()
	current := util.GetEnv(c)

	byt := b.NewBuffer([]byte{})

	w := tabwriter.NewWriter(byt, 0, 0, 1, ' ', 0)
	for i, env := range envs {
		if i > 0 {
			fmt.Fprintf(w, "\n")
		}
		prefix := " "
		if env.Name == current.Name {
			prefix = "*"
		}
		if env.ProxyAddress == "" {
			env.ProxyAddress = "none"
		}
		fmt.Fprintf(w, "%v %v \t %v", prefix, env.Name, env.ProxyAddress)
	}
	w.Flush()
	return byt.Bytes(), nil
}

func addEnv(c *cli.Context, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("name required")
	}
	if len(args) == 1 {
		args = append(args, "") // default to no proxy address
	}

	util.AddEnv(util.Env{
		Name:         args[0],
		ProxyAddress: args[1],
	})
	return nil, nil
}

// netCall calls services through the network
func netCall(c *cli.Context, args []string) ([]byte, error) {
	os.Setenv("VINE_PROXY", "go.vine.network")
	return clic.CallService(c, args)
}

// TODO: stream via HTTP
func streamService(c *cli.Context, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("require service and endpoint")
	}
	service := args[0]
	endpoint := args[1]
	var request map[string]interface{}

	// ignore error
	json.Unmarshal([]byte(strings.Join(args[2:], " ")), &request)

	req := (*cmd.DefaultOptions().Client).NewRequest(service, endpoint, request, client.WithContentType("application/json"))
	stream, err := (*cmd.DefaultOptions().Client).Stream(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("error calling %s.%s: %v", service, endpoint, err)
	}

	if err := stream.Send(request); err != nil {
		return nil, fmt.Errorf("error sending to %s.%s: %v", service, endpoint, err)
	}

	output := c.String("output")

	for {
		if output == "raw" {
			rsp := bytes.Frame{}
			if err := stream.Recv(&rsp); err != nil {
				return nil, fmt.Errorf("error receiving from %s.%s: %v", service, endpoint, err)
			}
			fmt.Print(string(rsp.Data))
		} else {
			var response map[string]interface{}
			if err := stream.Recv(&response); err != nil {
				return nil, fmt.Errorf("error receiving from %s.%s: %v", service, endpoint, err)
			}
			b, _ := json.MarshalIndent(response, "", "\t")
			fmt.Print(string(b))
		}
	}
}

func publish(c *cli.Context, args []string) ([]byte, error) {
	if err := clic.Publish(c, args); err != nil {
		return nil, err
	}
	return []byte(`ok`), nil
}

func queryHealth(c *cli.Context, args []string) ([]byte, error) {
	return clic.QueryHealth(c, args)
}

func queryStats(c *cli.Context, args []string) ([]byte, error) {
	return clic.QueryStats(c, args)
}

func upload(ctx *cli.Context, args []string) ([]byte, error) {
	if ctx.Args().Len() == 0 {
		return nil, errors.New("Required filename to upload")
	}

	filename := ctx.Args().Get(0)
	localfile := ctx.Args().Get(1)

	fileClient := file.New("go.vine.server", client.DefaultClient)
	return nil, fileClient.Upload(filename, localfile)
}
