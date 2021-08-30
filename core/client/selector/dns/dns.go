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

// Package dns provides a dns SRV selector
package dns

import (
	"fmt"
	"net"
	"strconv"

	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
)

type dnsSelector struct {
	options selector.Options
	domain  string
}

var (
	DefaultDomain = "local"
)

func (d *dnsSelector) Init(opts ...selector.Option) error {
	for _, o := range opts {
		o(&d.options)
	}
	return nil
}

func (d *dnsSelector) Options() selector.Options {
	return d.options
}

func (d *dnsSelector) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	var svc []*net.SRV

	// check if its host:port
	host, port, err := net.SplitHostPort(service)
	// not host:port
	if err != nil {
		// lookup the SRV record
		_, svcs, err := net.LookupSRV(service, "tcp", d.domain)
		if err != nil {
			return nil, err
		}
		// set SRV records
		svc = svcs
		// got host:port
	} else {
		p, _ := strconv.Atoi(port)

		// lookup the A record
		ips, err := net.LookupHost(host)
		if err != nil {
			return nil, err
		}

		// create SRV records
		for _, ip := range ips {
			svc = append(svc, &net.SRV{
				Target: ip,
				Port:   uint16(p),
			})
		}
	}

	nodes := make([]*registry.Node, 0, len(svc))
	for _, node := range svc {
		nodes = append(nodes, &registry.Node{
			Id:      node.Target,
			Address: fmt.Sprintf("%s:%d", node.Target, node.Port),
		})
	}

	services := []*registry.Service{
		{
			Name:  service,
			Nodes: nodes,
		},
	}

	sopts := selector.SelectOptions{
		Strategy: d.options.Strategy,
	}

	for _, opt := range opts {
		opt(&sopts)
	}

	// apply the filters
	for _, filter := range sopts.Filters {
		services = filter(services)
	}

	// if there's nothing left, return
	if len(services) == 0 {
		return nil, selector.ErrNoneAvailable
	}

	return sopts.Strategy(services), nil
}

func (d *dnsSelector) Mark(service string, node *registry.Node, err error) {}

func (d *dnsSelector) Reset(service string) {}

func (d *dnsSelector) Close() error {
	return nil
}

func (d *dnsSelector) String() string {
	return "dns"
}

func NewSelector(opts ...selector.Option) selector.Selector {
	options := selector.Options{
		Strategy: selector.Random,
	}

	for _, o := range opts {
		o(&options)
	}

	return &dnsSelector{options: options, domain: DefaultDomain}
}
