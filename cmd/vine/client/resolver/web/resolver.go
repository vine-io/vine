// MIT License
//
// Copyright (c) 2020 The vine Authors
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

package web

import (
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/vine-io/vine/core/client/selector"
	res "github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/util/namespace"
	"golang.org/x/net/publicsuffix"
)

var (
	re               = regexp.MustCompile("^[a-zA-Z0-9]+([a-zA-Z0-9-]*[a-zA-Z0-9]*)?$")
	defaultNamespace = namespace.DefaultNamespace + ".web"
)

type Resolver struct {
	// Type of resolver e.g path, domain
	Type string
	// a function which returns the namespace of the request
	Namespace func(req *http.Request) string
	// selector to find services
	Selector selector.Selector
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (r *Resolver) String() string {
	return "web/resolver"
}

// Info checks whether this is a web request.
// It returns host, namespace and whether its internal
func (r *Resolver) Info(req *http.Request) (string, string, bool) {
	// set to host
	host := req.URL.Hostname()

	// set as req.Host if blank
	if len(host) == 0 {
		host = req.Host
	}

	// split out ip
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// determine the namespace of the request
	namespace := r.Namespace(req)

	// overide host if the namespace is go.vine.web, since
	// this will also catch localhost & 127.0.0.1, resulting
	// in a more consistent dev experience
	if host == "localhost" || host == "127.0.0.1" {
		host = "web.vine.mu"
	}

	// if the type is path, always resolve using the path
	if r.Type == "path" {
		return host, namespace, true
	}

	// if the namespace is not the default (go.vine.web),
	// we always resolve using path
	if namespace != defaultNamespace {
		return host, namespace, true
	}

	// check for vine subdomains, we want to do subdomain routing
	// on these if the subdomoain routing has been specified
	if r.Type == "subdomain" && host != "web.vine.mu" && strings.HasSuffix(host, ".vine.mu") {
		return host, namespace, false
	}

	// Check for services info path, also handled by vine web but
	// not a top level path. TODO: Find a better way of detecting and
	// handling the non-proxied paths.
	if strings.HasPrefix(req.URL.Path, "/service/") {
		return host, namespace, true
	}

	// Check if the request is a top level path
	isWeb := strings.Count(req.URL.Path, "/") == 1
	return host, namespace, isWeb
}

// Resolve replaces the values of Host, Path, Scheme to calla backend service
// It accounts for subdomains for service names based on namespace
func (r *Resolver) Resolve(req *http.Request) (*res.Endpoint, error) {
	// get host, namespace and if its an internal request
	host, _, _ := r.Info(req)

	// check for vine web
	if r.Type == "path" || host == "web.vine.mu" {
		return r.resolveWithPath(req)
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return nil, err
	}

	// get and reverse the subdomain
	subdomain := strings.TrimSuffix(host, "."+domain)
	parts := strings.Split(subdomain, ".")
	reverse(parts)

	// turn it into an alias
	alias := strings.Join(parts, ".")
	if len(alias) == 0 {
		return nil, errors.New("unknown host")
	}

	var name string
	if strings.HasSuffix(host, ".vine.mu") {
		// for vine.mu subdomains, we route foo.vine.mu/bar to
		// go.vine.web.bar
		name = defaultNamespace + "." + alias
	} else if comps := strings.Split(req.URL.Path, "/"); len(comps) > 0 {
		// for non vine.mu subdomains, we route foo.m3o.app/bar to
		// foo.web.bar
		name = alias + ".web." + comps[1]
	}

	// find the service using the selector
	next, err := r.Selector.Select(name)
	if err == selector.ErrNotFound {
		// fallback to path based
		return r.resolveWithPath(req)
	} else if err != nil {
		return nil, err
	}

	// TODO: better retry strategy
	s, err := next()
	if err != nil {
		return nil, err
	}

	// we're done
	return &res.Endpoint{
		Name:   alias,
		Method: req.Method,
		Host:   s.Address,
		Path:   req.URL.Path,
	}, nil
}

func (r *Resolver) resolveWithPath(req *http.Request) (*res.Endpoint, error) {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 2 {
		return nil, errors.New("unknown service")
	}

	if !re.MatchString(parts[1]) {
		return nil, res.ErrInvalidPath
	}

	_, namespace, _ := r.Info(req)
	next, err := r.Selector.Select(namespace + "." + parts[1])
	if err == selector.ErrNotFound {
		return nil, res.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	// TODO: better retry strategy
	s, err := next()
	if err != nil {
		return nil, err
	}

	// we're done
	return &res.Endpoint{
		Name:   parts[1],
		Method: req.URL.Path,
		Host:   s.Address,
		Path:   "/" + strings.Join(parts[2:], "/"),
	}, nil
}
