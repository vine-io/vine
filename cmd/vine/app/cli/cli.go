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

// Package cli is a command line interface
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/lack-io/cli"

	storecli "github.com/lack-io/vine/lib/store/cli"

	"github.com/chzyer/readline"
)

var (
	prompt = "vine> "

	commands = map[string]*command{
		"quit":       {"quit", "Exit the CLI", quit},
		"exit":       {"exit", "Exit the CLI", quit},
		"call":       {"call", "Call a service", callService},
		"list":       {"list", "List services, peers or routes", list},
		"get":        {"get", "Get service info", getService},
		"stream":     {"stream", "Stream a call to a service", streamService},
		"publish":    {"publish", "Publish a message to a topic", publish},
		"health":     {"health", "Get service health", queryHealth},
		"stats":      {"stats", "Get service stats", queryStats},
		"register":   {"register", "Register a service", registerService},
		"deregister": {"deregister", "Deregister a service", deregisterService},
	}
)

type command struct {
	name  string
	usage string
	exec  exec
}

func Run(c *cli.Context) error {
	commands["help"] = &command{"help", "CLI usage", help}
	alias := map[string]string{
		"?":  "help",
		"ls": "list",
	}

	r, err := readline.New(prompt)
	if err != nil {
		// TODO return err
		fmt.Fprint(os.Stdout, err)
		os.Exit(1)
	}
	defer r.Close()

	for {
		args, err := r.Readline()
		if err != nil {
			fmt.Fprint(os.Stdout, err)
			return err
		}

		args = strings.TrimSpace(args)

		// skip no args
		if len(args) == 0 {
			continue
		}

		parts := strings.Split(args, " ")
		if len(parts) == 0 {
			continue
		}

		name := parts[0]

		// get alias
		if n, ok := alias[name]; ok {
			name = n
		}

		if cmd, ok := commands[name]; ok {
			rsp, err := cmd.exec(c, parts[1:])
			if err != nil {
				// TODO return err
				println(err.Error())
				continue
			}
			println(string(rsp))
		} else {
			// TODO return err
			println("unknown command")
		}
	}
	return nil
}

//NetworkCommands for network toplogy routing
func NetworkCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "connect",
			Usage:  "connect to the network. specify nodes e.g connect ip:port",
			Action: Print(networkConnect),
		},
		{
			Name:   "connections",
			Usage:  "List the immediate connections to the network",
			Action: Print(networkConnections),
		},
		{
			Name:   "graph",
			Usage:  "Get the network graph",
			Action: Print(networkGraph),
		},
		{
			Name:   "nodes",
			Usage:  "List nodes in the network",
			Action: Print(netNodes),
		},
		{
			Name:   "routes",
			Usage:  "List network routes",
			Action: Print(netRoutes),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "service",
					Usage: "Filter by service",
				},
				&cli.StringFlag{
					Name:  "address",
					Usage: "Filter by address",
				},
				&cli.StringFlag{
					Name:  "gateway",
					Usage: "Filter by gateway",
				},
				&cli.StringFlag{
					Name:  "router",
					Usage: "Filter by router",
				},
				&cli.StringFlag{
					Name:  "network",
					Usage: "Filter by network",
				},
			},
		},
		{
			Name:   "services",
			Usage:  "Get the network services",
			Action: Print(networkServices),
		},
		// TODO: duplicates call. Move so we reuse same stuff.
		{
			Name:   "call",
			Usage:  "Call a service e.g vine call greeter Say.Hello '{\"name\": \"John\"}",
			Action: Print(netCall),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Set the address of the service instance to call",
					EnvVars: []string{"VINE_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "output, o",
					Usage:   "Set the output format; json (default), raw",
					EnvVars: []string{"VINE_OUTPUT"},
				},
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
			},
		},
	}
}

//NetworkDNSCommands for networking routing
func NetworkDNSCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "advertise",
			Usage: "Advertise a new node to the network",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Address to register for the specified domain",
					EnvVars: []string{"VINE_NETWORK_DNS_ADVERTISE_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "domain",
					Usage:   "Domain name to register",
					EnvVars: []string{"VINE_NETWORK_DNS_ADVERTISE_DOMAIN"},
					Value:   "network.vine.mu",
				},
				&cli.StringFlag{
					Name:    "token",
					Usage:   "Bearer token for the go.vine.network.dns service",
					EnvVars: []string{"VINE_NETWORK_DNS_ADVERTISE_TOKEN"},
				},
			},
			Action: Print(netDNSAdvertise),
		},
		{
			Name:  "remove",
			Usage: "Remove a node's record'",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Address to register for the specified domain",
					EnvVars: []string{"VINE_NETWORK_DNS_REMOVE_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "domain",
					Usage:   "Domain name to remove",
					EnvVars: []string{"VINE_NETWORK_DNS_REMOVE_DOMAIN"},
					Value:   "network.vine.mu",
				},
				&cli.StringFlag{
					Name:    "token",
					Usage:   "Bearer token for the go.vine.network.dns service",
					EnvVars: []string{"VINE_NETWORK_DNS_REMOVE_TOKEN"},
				},
			},
			Action: Print(netDNSRemove),
		},
		{
			Name:  "resolve",
			Usage: "Remove a record'",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "domain",
					Usage:   "Domain name to resolve",
					EnvVars: []string{"VINE_NETWORK_DNS_RESOLVE_DOMAIN"},
					Value:   "network.vine.mu",
				},
				&cli.StringFlag{
					Name:    "type",
					Usage:   "Domain name type to resolve",
					EnvVars: []string{"VINE_NETWORK_DNS_RESOLVE_TYPE"},
					Value:   "A",
				},
				&cli.StringFlag{
					Name:    "token",
					Usage:   "Bearer token for the go.vine.network.dns service",
					EnvVars: []string{"VINE_NETWORK_DNS_RESOLVE_TOKEN"},
				},
			},
			Action: Print(netDNSResolve),
		},
	}
}

func RegistryCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "list",
			Usage: "List items in registry or network",
			Subcommands: []*cli.Command{
				{
					Name:   "nodes",
					Usage:  "List nodes in the network",
					Action: Print(netNodes),
				},
				{
					Name:   "routes",
					Usage:  "List network routes",
					Action: Print(netRoutes),
				},
				{
					Name:   "services",
					Usage:  "List services in registry",
					Action: Print(listServices),
				},
			},
		},
		{
			Name:  "register",
			Usage: "Register an item in the registry",
			Subcommands: []*cli.Command{
				{
					Name:   "service",
					Usage:  "Register a service with JSON definition",
					Action: Print(registerService),
				},
			},
		},
		{
			Name:  "deregister",
			Usage: "Deregister an item in the registry",
			Subcommands: []*cli.Command{
				{
					Name:   "service",
					Usage:  "Deregister a service with JSON definition",
					Action: Print(deregisterService),
				},
			},
		},
		{
			Name:  "get",
			Usage: "Get item from registry",
			Subcommands: []*cli.Command{
				{
					Name:   "service",
					Usage:  "Get service from registry",
					Action: Print(getService),
				},
			},
		},
	}
}

//StoreCommands for data storing
func StoreCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "snapshot",
			Usage:  "Back up a store",
			Action: storecli.Snapshot,
			Flags: append(storecli.CommonFlags,
				&cli.StringFlag{
					Name:    "destination",
					Usage:   "Backup destination",
					Value:   "file:///tmp/store-snapshot",
					EnvVars: []string{"VINE_SNAPSHOT_DESTINATION"},
				},
			),
		},
		{
			Name:   "sync",
			Usage:  "Copy all records of one store into another store",
			Action: storecli.Sync,
			Flags:  storecli.SyncFlags,
		},
		{
			Name:   "restore",
			Usage:  "restore a store snapshot",
			Action: storecli.Restore,
			Flags: append(storecli.CommonFlags,
				&cli.StringFlag{
					Name:  "source",
					Usage: "Backup source",
					Value: "file:///tmp/store-snapshot",
				},
			),
		},
		{
			Name:   "databases",
			Usage:  "List all databases known to the store service",
			Action: storecli.Databases,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "store",
					Usage: "store service to call",
					Value: "go.vine.store",
				},
			},
		},
		{
			Name:   "tables",
			Usage:  "List all tables in the specified database known to the store service",
			Action: storecli.Tables,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "store",
					Usage: "store service to call",
					Value: "go.vine.store",
				},
				&cli.StringFlag{
					Name:    "database",
					Aliases: []string{"d"},
					Usage:   "database to list tables of",
					Value:   "vine",
				},
			},
		},
		{
			Name:      "read",
			Usage:     "read a record from the store",
			UsageText: `vine store read [options] key`,
			Action:    storecli.Read,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "database",
					Aliases: []string{"d"},
					Usage:   "database to write to",
					Value:   "vine",
				},
				&cli.StringFlag{
					Name:    "table",
					Aliases: []string{"t"},
					Usage:   "table to write to",
					Value:   "vine",
				},
				&cli.BoolFlag{
					Name:    "prefix",
					Aliases: []string{"p"},
					Usage:   "read prefix",
					Value:   false,
				},
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
					Usage:   "show keys and headers (only values shown by default)",
					Value:   false,
				},
				&cli.StringFlag{
					Name:  "output",
					Usage: "output format (json, table)",
					Value: "table",
				},
			},
		},
		{
			Name:      "list",
			Usage:     "list all keys from a store",
			UsageText: `vine store list [options]`,
			Action:    storecli.List,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "database",
					Aliases: []string{"d"},
					Usage:   "database to list from",
					Value:   "vine",
				},
				&cli.StringFlag{
					Name:    "table",
					Aliases: []string{"t"},
					Usage:   "table to write to",
					Value:   "vine",
				},
				&cli.StringFlag{
					Name:  "output",
					Usage: "output format (json)",
				},
				&cli.BoolFlag{
					Name:    "prefix",
					Aliases: []string{"p"},
					Usage:   "list prefix",
					Value:   false,
				},
				&cli.UintFlag{
					Name:    "limit",
					Aliases: []string{"l"},
					Usage:   "list limit",
				},
				&cli.UintFlag{
					Name:    "offset",
					Aliases: []string{"o"},
					Usage:   "list offset",
				},
			},
		},
		{
			Name:      "write",
			Usage:     "write a record to the store",
			UsageText: `vine store write [options] key value`,
			Action:    storecli.Write,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "expiry",
					Aliases: []string{"e"},
					Usage:   "expiry in time.ParseDuration format",
					Value:   "",
				},
				&cli.StringFlag{
					Name:    "database",
					Aliases: []string{"d"},
					Usage:   "database to write to",
					Value:   "vine",
				},
				&cli.StringFlag{
					Name:    "table",
					Aliases: []string{"t"},
					Usage:   "table to write to",
					Value:   "vine",
				},
			},
		},
		{
			Name:      "delete",
			Usage:     "delete a key from the store",
			UsageText: `vine store delete [options] key`,
			Action:    storecli.Delete,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "database",
					Usage: "database to delete from",
					Value: "vine",
				},
				&cli.StringFlag{
					Name:  "table",
					Usage: "table to delete from",
					Value: "vine",
				},
			},
		},
	}
}

//Commands for vine calling action
func Commands() []*cli.Command {
	commands := []*cli.Command{
		{
			Name:   "cli",
			Usage:  "Run the interactive CLI",
			Action: Run,
		},
		{
			Name:   "call",
			Usage:  "Call a service e.g vine call greeter Say.Hello '{\"name\": \"John\"}",
			Action: Print(callService),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Set the address of the service instance to call",
					EnvVars: []string{"VINE_ADDRESS"},
				},
				&cli.StringFlag{
					Name:    "output, o",
					Usage:   "Set the output format; json (default), raw",
					EnvVars: []string{"VINE_OUTPUT"},
				},
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
			},
		},
		{
			Name:   "stream",
			Usage:  "Create a service stream",
			Action: Print(streamService),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output, o",
					Usage:   "Set the output format; json (default), raw",
					EnvVars: []string{"VINE_OUTPUT"},
				},
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "Publish a message to a topic",
			Action: Print(publish),
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "metadata",
					Usage:   "A list of key-value pairs to be forwarded as metadata",
					EnvVars: []string{"VINE_METADATA"},
				},
			},
		},
		{
			Name:   "stats",
			Usage:  "Query the stats of a service",
			Action: Print(queryStats),
		},
		{
			Name:   "env",
			Usage:  "Get/set vine cli environment",
			Action: Print(listEnvs),
			Subcommands: []*cli.Command{
				{
					Name:   "get",
					Action: Print(getEnv),
				},
				{
					Name:   "set",
					Action: Print(setEnv),
				},
				{
					Name:   "add",
					Action: Print(addEnv),
				},
			},
		},
		{
			Name:  "file",
			Usage: "Move files between your local machine and the server",
			Subcommands: []*cli.Command{
				{
					Name:   "upload",
					Action: Print(upload),
				},
			},
		},
	}

	return append(commands, RegistryCommands()...)
}
