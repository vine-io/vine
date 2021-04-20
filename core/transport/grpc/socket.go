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

package grpc

import (
	"google.golang.org/grpc"

	"github.com/lack-io/vine/core/transport"
	pb "github.com/lack-io/vine/proto/services/transport"
)

type grpcTransportClient struct {
	conn   *grpc.ClientConn
	stream pb.Transport_StreamClient

	local  string
	remote string
}

type grpcTransportSocket struct {
	stream pb.Transport_StreamServer
	local  string
	remote string
}

func (g *grpcTransportClient) Local() string {
	return g.local
}

func (g *grpcTransportClient) Remote() string {
	return g.remote
}

func (g *grpcTransportClient) Recv(m *transport.Message) error {
	if m == nil {
		return nil
	}

	msg, err := g.stream.Recv()
	if err != nil {
		return err
	}

	m.Header = msg.Header
	m.Body = msg.Body
	return nil
}

func (g *grpcTransportClient) Send(m *transport.Message) error {
	if m == nil {
		return nil
	}

	return g.stream.Send(&pb.Message{
		Header: m.Header,
		Body:   m.Body,
	})
}

func (g *grpcTransportClient) Close() error {
	return g.conn.Close()
}

func (g *grpcTransportSocket) Local() string {
	return g.local
}

func (g *grpcTransportSocket) Remote() string {
	return g.remote
}

func (g *grpcTransportSocket) Recv(m *transport.Message) error {
	if m == nil {
		return nil
	}

	msg, err := g.stream.Recv()
	if err != nil {
		return err
	}

	m.Header = msg.Header
	m.Body = msg.Body
	return nil
}

func (g *grpcTransportSocket) Send(m *transport.Message) error {
	if m == nil {
		return nil
	}

	return g.stream.Send(&pb.Message{
		Header: m.Header,
		Body:   m.Body,
	})
}

func (g *grpcTransportSocket) Close() error {
	return nil
}
