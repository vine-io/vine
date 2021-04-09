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

package tunnel

import (
	"io"
	"sync"

	"github.com/lack-io/vine/service/logger"
)

type tunListener struct {
	// address of the listener
	channel string
	// token is the tunnel token
	token string
	// the accept channel
	accept chan *session
	// the tunnel closed channel
	tunClosed chan bool
	// the listener session
	session *session
	// del func to kill listener
	delFunc func()

	sync.RWMutex
	// the channel to close
	closed chan bool
}

func (t *tunListener) process() {
	// our connection map for session
	conns := make(map[string]*session)

	defer func() {
		// close the sessions
		for id, conn := range conns {
			conn.Close()
			delete(conns, id)
		}
		// unassign
		conns = nil
	}()

	for {
		select {
		case <-t.closed:
			return
		case <-t.tunClosed:
			t.Close()
			return
		// receive a new message
		case m := <-t.session.recv:
			var sessionId string
			var linkId string

			switch t.session.mode {
			case Multicast:
				sessionId = "multicast"
				linkId = "multicast"
			case Broadcast:
				sessionId = "broadcast"
				linkId = "broadcast"
			default:
				sessionId = m.session
				linkId = m.link
			}

			// get a session
			sess, ok := conns[sessionId]
			logger.Tracef("Tunnel listener received channel %+v session %s type %s exists: %t", m.channel, m.session, m.typ, ok)
			if !ok {
				// we only process open and session types
				switch m.typ {
				case "open", "session":
				default:
					continue
				}

				// create a new session session
				sess = &session{
					// the session key
					key: []byte(t.token + m.channel + sessionId),
					// the id of the remote side
					tunnel: m.tunnel,
					// the channel
					channel: m.channel,
					// the session id
					session: sessionId,
					// tunnel token
					token: t.token,
					// is loopback conn
					loopback: m.loopback,
					// the link the message was received on
					link: linkId,
					// set the connection mode
					mode: t.session.mode,
					// close chan
					closed: make(chan bool),
					// recv called by the acceptor
					recv: make(chan *message, 128),
					// use the internal send buffer
					send: t.session.send,
					// error channel
					errChan: make(chan error, 1),
					// set the read timeout
					readTimeout: t.session.readTimeout,
				}

				// save the session
				conns[sessionId] = sess

				select {
				case <-t.closed:
					return
				// send to accept chan
				case t.accept <- sess:
				}
			}

			// an existing session was found

			switch m.typ {
			case "close":
				// don't close multicast sessions
				if sess.mode > Unicast {
					continue
				}

				// received a close message
				select {
				// check if the session is closed
				case <-sess.closed:
					// no op
					delete(conns, sessionId)
				default:
					// only close if unicast session
					// close and delete session
					close(sess.closed)
					delete(conns, sessionId)
				}

				// continue
				continue
			case "session":
				// operate on this
			default:
				// non operational type
				continue
			}

			// send this to the accept chan
			select {
			case <-sess.closed:
				delete(conns, sessionId)
			case sess.recv <- m:
				logger.Tracef("Tunnel listener sent to recv chan channel %+v session %s type %s", m.channel, sessionId, m.typ)
			}
		}
	}
}

func (t *tunListener) Channel() string {
	return t.channel
}

// Close closes tunnel listener
func (t *tunListener) Close() error {
	t.Lock()
	defer t.Unlock()

	select {
	case <-t.closed:
		return nil
	default:
		// close and delete
		t.delFunc()
		t.session.Close()
		close(t.closed)
	}
	return nil
}

// Everytime accept is called we essentially block till we get a new connection
func (t *tunListener) Accept() (Session, error) {
	select {
	// if the session is closed return
	case <-t.closed:
		return nil, io.EOF
	case <-t.tunClosed:
		// close the listener when the tunnel closes
		return nil, io.EOF
	// wait for a new connection
	case c, ok := <-t.accept:
		// check if the accept chan is closed
		if !ok {
			return nil, io.EOF
		}
		// return without accept
		if c.mode != Unicast {
			return c, nil
		}
		// send back the accept
		if err := c.Accept(); err != nil {
			return nil, err
		}
		return c, nil
	}
}
