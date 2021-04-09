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

package pool

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/lack-io/vine/service/transport"
)

type pool struct {
	size int
	ttl  time.Duration
	tr   transport.Transport

	sync.Mutex
	conns map[string][]*poolConn
}

type poolConn struct {
	transport.Client
	id      string
	created time.Time
}

func newPool(options Options) *pool {
	return &pool{
		size:  options.Size,
		tr:    options.Transport,
		ttl:   options.TTL,
		conns: make(map[string][]*poolConn),
	}
}

func (p *pool) Close() error {
	p.Lock()
	for k, c := range p.conns {
		for _, conn := range c {
			conn.Client.Close()
		}
		delete(p.conns, k)
	}
	p.Unlock()
	return nil
}

// NoOp the Close since we manage it
func (p *poolConn) Close() error {
	return nil
}

func (p *poolConn) Id() string {
	return p.id
}

func (p *poolConn) Created() time.Time {
	return p.created
}

func (p *pool) Get(addr string, opts ...transport.DialOption) (Conn, error) {
	p.Lock()
	conns := p.conns[addr]

	// while we have conns check age and then return one
	// otherwise we'll create a new conn
	for len(conns) > 0 {
		conn := conns[len(conns)-1]
		conns = conns[:len(conns)-1]
		p.conns[addr] = conns

		// if conn is old kill it and move on
		if d := time.Since(conn.created); d > p.ttl {
			conn.Client.Close()
			continue
		}

		// we got a good conn, lets unlock and return it
		p.Unlock()

		return conn, nil
	}

	p.Unlock()

	// create new conn
	c, err := p.tr.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &poolConn{
		Client:  c,
		id:      uuid.New().String(),
		created: time.Now(),
	}, nil
}

func (p *pool) Release(conn Conn, err error) error {
	// don't store the conn if it has errored
	if err != nil {
		return conn.(*poolConn).Client.Close()
	}

	// otherwise put it back ofr reuse
	p.Lock()
	conns := p.conns[conn.Remote()]
	if len(conns) >= p.size {
		p.Unlock()
		return conn.(*poolConn).Client.Close()
	}
	p.conns[conn.Remote()] = append(conns, conn.(*poolConn))
	p.Unlock()

	return nil
}
