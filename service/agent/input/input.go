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
	// Flags Provide cli flags
	Flags() []cli.Flag
	// Init Initialise input using cli context
	Init(*cli.Context) error
	// Stream events from the input
	Stream() (Conn, error)
	// Start the input
	Start() error
	// Stop the input
	Stop() error
	// String name of the input
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
