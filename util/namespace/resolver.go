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

package namespace

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/publicsuffix"

	log "github.com/lack-io/vine/lib/logger"
)

func NewResolver(svcType, namespace string) *Resolver {
	return &Resolver{svcType, namespace}
}

// Resolver determines the namespace for a request
type Resolver struct {
	svcType   string
	namespace string
}

func (r Resolver) String() string {
	return "internal/namespace"
}

func (r Resolver) ResolveWithType(c *fiber.Ctx) string {
	return r.Resolve(c) + "." + r.svcType
}

func (r Resolver) Resolve(c *fiber.Ctx) string {
	// check to see what the provided namespace is, we only do
	// domain mapping if the namespace is set to 'domain'
	if r.namespace != "domain" {
		return r.namespace
	}

	// determine the host, e.g. dev.vine.mu:8080
	host := c.Hostname()
	if len(host) == 0 {
		if h, _, err := net.SplitHostPort(string(c.Request().Host())); err == nil {
			host = h // host does contain a port
		} else if strings.Contains(err.Error(), "missing port in address") {
			host = string(c.Request().Host()) // host does not contain a port
		}
	}

	// check for an ip address
	if net.ParseIP(host) != nil {
		return DefaultNamespace
	}

	// check for dev enviroment
	if host == "localhost" || host == "127.0.0.1" {
		return DefaultNamespace
	}

	// extract the top level domain plus one (e.g. 'myapp.com')
	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		log.Debugf("Unable to extract domain from %v", host)
		return DefaultNamespace
	}

	// check to see if the domain matches the host of vine.mu, in
	// these cases we return the default namespace
	if domain == host || domain == "vine.mu" {
		return DefaultNamespace
	}

	// remove the domain from the host, leaving the subdomain
	subdomain := strings.TrimSuffix(host, "."+domain)

	// return the reversed subdomain as the namespace
	comps := strings.Split(subdomain, ".")
	for i := len(comps)/2 - 1; i >= 0; i-- {
		opp := len(comps) - 1 - i
		comps[i], comps[opp] = comps[opp], comps[i]
	}
	return strings.Join(comps, ".")
}
