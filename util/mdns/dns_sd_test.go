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

package mdns

import (
	"reflect"
	"testing"

	"github.com/miekg/dns"
)

type mockMDNSService struct{}

func (s *mockMDNSService) Records(q dns.Question) []dns.RR {
	return []dns.RR{
		&dns.PTR{
			Hdr: dns.RR_Header{
				Name:   "fakerecord",
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    42,
			},
			Ptr: "fake.local.",
		},
	}
}

func (s *mockMDNSService) Announcement() []dns.RR {
	return []dns.RR{
		&dns.PTR{
			Hdr: dns.RR_Header{
				Name:   "fakeannounce",
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    42,
			},
			Ptr: "fake.local.",
		},
	}
}

func TestDNSSDServiceRecords(t *testing.T) {
	s := &DNSSDService{
		MDNSService: &MDNSService{
			serviceAddr: "_foobar._tcp.local.",
			Domain:      "local",
		},
	}
	q := dns.Question{
		Name:   "_services._dns-sd._udp.local.",
		Qtype:  dns.TypePTR,
		Qclass: dns.ClassINET,
	}
	recs := s.Records(q)
	if got, want := len(recs), 1; got != want {
		t.Fatalf("s.Records(%v) returned %v records, want %v", q, got, want)
	}

	want := dns.RR(&dns.PTR{
		Hdr: dns.RR_Header{
			Name:   "_services._dns-sd._udp.local.",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    defaultTTL,
		},
		Ptr: "_foobar._tcp.local.",
	})
	if got := recs[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("s.Records()[0] = %v, want %v", got, want)
	}
}
