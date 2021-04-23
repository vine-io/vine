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
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/core/registry/memory"
	"github.com/lack-io/vine/lib/agent/command"
	"github.com/lack-io/vine/lib/agent/input"
)

type testInput struct {
	send chan *input.Event
	recv chan *input.Event
	exit chan bool
}

func (t *testInput) Flags() []cli.Flag {
	return nil
}

func (t *testInput) Init(*cli.Context) error {
	return nil
}

func (t *testInput) Close() error {
	select {
	case <-t.exit:
	default:
		close(t.exit)
	}
	return nil
}

func (t *testInput) Send(event *input.Event) error {
	if event == nil {
		return errors.New("nil event")
	}

	select {
	case <-t.exit:
		return errors.New("connection closed")
	case t.send <- event:
		return nil
	}
}

func (t *testInput) Recv(event *input.Event) error {
	if event == nil {
		return errors.New("nil event")
	}

	select {
	case <-t.exit:
		return errors.New("connection closed")
	case ev := <-t.recv:
		*event = *ev
		return nil
	}

}

func (t *testInput) Start() error {
	return nil
}

func (t *testInput) Stop() error {
	return nil
}

func (t *testInput) Stream() (input.Conn, error) {
	return t, nil
}

func (t *testInput) String() string {
	return "test"
}

func TestBot(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ExitOnError)
	app := cli.NewApp()
	ctx := cli.NewContext(app, flagSet, nil)

	io := &testInput{
		send: make(chan *input.Event),
		recv: make(chan *input.Event),
		exit: make(chan bool),
	}

	inputs := map[string]input.Input{
		"test": io,
	}

	commands := map[string]command.Command{
		"^echo ": command.NewCommand("echo", "test usage", "test description", func(args ...string) ([]byte, error) {
			return []byte(strings.Join(args[1:], " ")), nil
		}),
	}

	svc := vine.NewService(
		vine.Registry(memory.NewRegistry()),
	)

	bot := newBot(ctx, inputs, commands, svc)

	if err := bot.start(); err != nil {
		t.Fatal(err)
	}

	// send command
	select {
	case io.recv <- &input.Event{
		Meta: map[string]interface{}{},
		Type: input.TextEvent,
		Data: []byte("echo test"),
	}:
	case <-time.After(time.Second):
		t.Fatal("timed out sending event")
	}

	// recv output
	select {
	case ev := <-io.send:
		if string(ev.Data) != "test" {
			t.Fatal("expected 'test', got: ", string(ev.Data))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out receiving event")
	}

	if err := bot.stop(); err != nil {
		t.Fatal(err)
	}
}
