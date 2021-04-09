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

// Package dns provides a DNS registration service for autodiscovery of core network nodes.
package dns

import (
	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/cmd/vine/service/network/dns/handler"
	"github.com/lack-io/vine/cmd/vine/service/network/dns/provider/cloudflare"
	dns "github.com/lack-io/vine/proto/services/network/dns"
	log "github.com/lack-io/vine/service/logger"
)

// Run is the entrypoint for network/dns
func Run(c *cli.Context) {

	if c.String("provider") != "cloudflare" {
		log.Fatal("The only implemented DNS provider is cloudflare")
	}

	dnsService := vine.NewService(
		vine.Name("go.vine.network.dns"),
	)

	// Create handler
	provider, err := cloudflare.New(c.String("api-token"), c.String("zone-id"))
	if err != nil {
		log.Fatal(err)
	}
	h := handler.New(
		provider,
		c.String("token"),
	)

	// Register Handler
	dns.RegisterDnsHandler(dnsService.Server(), h)

	// Run service
	if err := dnsService.Run(); err != nil {
		log.Fatal(err)
	}

}
