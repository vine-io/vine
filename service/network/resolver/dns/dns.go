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

// Package dns resolves names to dns records
package dns

import (
	"context"
	"net"

	"github.com/miekg/dns"

	"github.com/lack-io/vine/service/network/resolver"
)

// Resolver is a DNS network resolve
type Resolver struct {
	// The resolver address to use
	Address string
}

// Resolve assumes ID is a domain name e.g vine.mu
func (r *Resolver) Resolve(name string) ([]*resolver.Record, error) {
	host, port, err := net.SplitHostPort(name)
	if err != nil {
		host = name
		port = "8085"
	}

	if len(host) == 0 {
		host = "localhost"
	}

	if len(r.Address) == 0 {
		r.Address = "1.0.0.1:53"
	}

	//nolint:prealloc
	var records []*resolver.Record

	// parsed an actual ip
	if v := net.ParseIP(host); v != nil {
		records = append(records, &resolver.Record{
			Address: net.JoinHostPort(host, port),
		})
		return records, nil
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)
	rec, err := dns.ExchangeContext(context.Background(), m, r.Address)
	if err != nil {
		return nil, err
	}

	for _, answer := range rec.Answer {
		h := answer.Header()
		// check record type matches
		if h.Rrtype != dns.TypeA {
			continue
		}

		arec, _ := answer.(*dns.A)
		addr := arec.A.String()

		// join resolved record with port
		address := net.JoinHostPort(addr, port)
		// append to record set
		records = append(records, &resolver.Record{
			Address: address,
		})
	}

	// no records returned so just best effort it
	if len(records) == 0 {
		records = append(records, &resolver.Record{
			Address: net.JoinHostPort(host, port),
		})
	}

	return records, nil
}
