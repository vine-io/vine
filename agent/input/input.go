// Copyright 2020 The vine Authors
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

// Package input is an interface for bot inputs
package input

import (
	"github.com/lack-io/cli"
)

type EventType string

const (
	TextEvent EventType = "text"
)

var (
	// Inputs keyed by name
	// Example slack or hipchat
	Inputs = map[string]Input{}
)

// Event is the unit sent and received
type Event struct {
	Type EventType
	From string
	To   string
	Data []byte
	Meta map[string]interface{}
}

// Input is an interface for sources which
// provide a way to communicate with the bot.
// Slack, HipChat, XMPP, etc.
type Input interface {
	// Provide cli flags
	Flags() []cli.Flag
	// Initialise input using cli context
	Init(*cli.Context) error
	// Stream events from the input
	Stream() (Conn, error)
	// Start the input
	Start() error
	// Stop the input
	Stop() error
	// name of the input
	String() string
}

// Conn interface provides a way to
// send and receive events. Send and
// Recv both block until succeeding
// or failing.
type Conn interface {
	Close() error
	Recv(*Event) error
	Send(*Event) error
}
