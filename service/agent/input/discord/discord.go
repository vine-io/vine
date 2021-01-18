// Copyright 2020 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discord

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service/agent/input"
)

func init() {
	input.Inputs["discord"] = newInput()
}

func newInput() *discordInput {
	return &discordInput{}
}

type discordInput struct {
	token     string
	whitelist []string
	prefix    string
	prefixfn  func(string) (string, bool)
	botID     string

	session *discordgo.Session

	sync.Mutex
	running bool
	exit    chan struct{}
}

func (d *discordInput) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "discord-token",
			EnvVars: []string{"VINE_DISCORD_TOKEN"},
			Usage:   "Discord token (prefix with Bot if it's for bot account)",
		},
		&cli.StringFlag{
			Name:    "discord-whitelist",
			EnvVars: []string{"VINE_DISCORD_WHITELIST"},
			Usage:   "Discord Whitelist (seperated by ,)",
		},
		&cli.StringFlag{
			Name:    "discord-prefix",
			Usage:   "Discord Prefix",
			EnvVars: []string{"VINE_DISCORD_PREFIX"},
			Value:   "VINE ",
		},
	}
}

func (d *discordInput) Init(ctx *cli.Context) error {
	token := ctx.String("discord-token")
	whitelist := ctx.String("discord-whitelist")
	prefix := ctx.String("discord-prefix")

	if len(token) == 0 {
		return errors.New("require token")
	}

	d.token = token
	d.prefix = prefix

	if len(whitelist) > 0 {
		d.whitelist = strings.Split(whitelist, ",")
	}

	return nil
}

func (d *discordInput) Start() error {
	if len(d.token) == 0 {
		return errors.New("missing discord configuration")
	}

	d.Lock()
	defer d.Unlock()

	if d.running {
		return nil
	}

	var err error
	d.session, err = discordgo.New("Bot " + d.token)
	if err != nil {
		return err
	}

	u, err := d.session.User("@me")
	if err != nil {
		return err
	}

	d.botID = u.ID
	d.prefixfn = CheckPrefixFactory(fmt.Sprintf("<@%s> ", d.botID), fmt.Sprintf("<@!%s> ", d.botID), d.prefix)

	d.exit = make(chan struct{})
	d.running = true

	return nil
}

func (d *discordInput) Stream() (input.Conn, error) {
	d.Lock()
	defer d.Unlock()
	if !d.running {
		return nil, errors.New("not running")
	}

	//Fire-n-forget close just in case...
	d.session.Close()

	conn := newConn(d)
	if err := d.session.Open(); err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *discordInput) Stop() error {
	d.Lock()
	defer d.Unlock()

	if !d.running {
		return nil
	}

	close(d.exit)
	d.running = false
	return nil
}

func (d *discordInput) String() string {
	return "discord"
}

// CheckPrefixFactory Creates a prefix checking function and stuff.
func CheckPrefixFactory(prefixes ...string) func(string) (string, bool) {
	return func(content string) (string, bool) {
		for _, prefix := range prefixes {
			if strings.HasPrefix(content, prefix) {
				return strings.TrimPrefix(content, prefix), true
			}
		}
		return "", false
	}
}
