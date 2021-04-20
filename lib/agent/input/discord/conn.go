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

package discord

import (
	"errors"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/lack-io/vine/lib/agent/input"
	"github.com/lack-io/vine/lib/logger"
)

type discordConn struct {
	master *discordInput
	exit   chan struct{}
	recv   chan *discordgo.Message

	sync.Mutex
}

func newConn(master *discordInput) *discordConn {
	conn := &discordConn{
		master: master,
		exit:   make(chan struct{}),
		recv:   make(chan *discordgo.Message),
	}

	conn.master.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == master.botID {
			return
		}

		whitelisted := false
		for _, ID := range conn.master.whitelist {
			if m.Author.ID == ID {
				whitelisted = true
				break
			}
		}

		if len(master.whitelist) > 0 && !whitelisted {
			return
		}

		var valid bool
		m.Message.Content, valid = conn.master.prefixfn(m.Message.Content)
		if !valid {
			return
		}

		conn.recv <- m.Message
	})

	return conn
}

func (dc *discordConn) Recv(event *input.Event) error {
	for {
		select {
		case <-dc.exit:
			return errors.New("connection closed")
		case msg := <-dc.recv:

			event.From = msg.ChannelID + ":" + msg.Author.ID
			event.To = dc.master.botID
			event.Type = input.TextEvent
			event.Data = []byte(msg.Content)
			return nil
		}
	}
}

func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func (dc *discordConn) Send(e *input.Event) error {
	fields := strings.Split(e.To, ":")
	for _, chunk := range ChunkString(string(e.Data), 2000) {
		_, err := dc.master.session.ChannelMessageSend(fields[0], chunk)
		if err != nil {
			logger.Error("[bot][loop][send]", err)

		}
	}
	return nil
}

func (dc *discordConn) Close() error {
	if err := dc.master.session.Close(); err != nil {
		return err
	}

	select {
	case <-dc.exit:
		return nil
	default:
		close(dc.exit)
	}
	return nil
}
