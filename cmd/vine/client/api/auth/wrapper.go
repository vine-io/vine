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

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lack-io/vine/service/api/resolver"
	"github.com/lack-io/vine/service/api/server"
	"github.com/lack-io/vine/service/auth"
	log "github.com/lack-io/vine/service/logger"
	inauth "github.com/lack-io/vine/util/auth"
	ctx "github.com/lack-io/vine/util/context"
	"github.com/lack-io/vine/util/namespace"
)

// Wrapper wraps a handler and authenticates requests
func Wrapper(r resolver.Resolver, nr *namespace.Resolver) server.Wrapper {
	return func(h http.Handler) http.Handler {
		return authWrapper{
			handler:    h,
			resolver:   r,
			nsResolver: nr,
			auth:       auth.DefaultAuth,
		}
	}
}

type authWrapper struct {
	handler    http.Handler
	auth       auth.Auth
	resolver   resolver.Resolver
	nsResolver *namespace.Resolver
}

func (a authWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Determine the namespace and set it in the header
	ns := req.Header.Get(namespace.NamespaceKey)
	if len(ns) == 0 {
		ns = a.nsResolver.Resolve(req)
		req.Header.Set(namespace.NamespaceKey, ns)
	}

	// Set the metadata so we can access it in vine api / web
	req = req.WithContext(ctx.FromRequest(req))

	// Extract the token from the request
	var token string
	if header := req.Header.Get("Authorization"); len(header) > 0 {
		// Extract the auth token from the request
		if strings.HasPrefix(header, auth.BearerScheme) {
			token = header[len(auth.BearerScheme):]
		}
	} else {
		// Get the token out the cookies if not provided in headers
		if c, err := req.Cookie("vine-token"); err == nil && c != nil {
			token = strings.TrimPrefix(c.Value, inauth.TokenCookieName+"=")
			req.Header.Set("Authorization", auth.BearerScheme+token)
		}
	}

	// Get the account using the token, some are unauthenticated, so the lack of an
	// account doesn't necesserially mean a forbidden request
	acc, _ := a.auth.Inspect(token)

	// Ensure the accounts issuer matches the namespace being requested
	if acc != nil && len(acc.Issuer) > 0 && acc.Issuer != ns {
		http.Error(w, "Account not issued by "+ns, 403)
		return
	}

	// Determine the name of the service being requested
	endpoint, err := a.resolver.Resolve(req)
	if err == resolver.ErrInvalidPath || err == resolver.ErrNotFound {
		// a file not served by the resolver has been requested (e.g. favicon.ico)
		endpoint = &resolver.Endpoint{Path: req.URL.Path}
	} else if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
		return
	} else {
		// set the endpoint in the context so it can be used to resolve
		// the request later
		ctx := context.WithValue(req.Context(), resolver.Endpoint{}, endpoint)
		*req = *req.Clone(ctx)
	}

	// construct the resource name, e.g. home => go.vine.web.home
	resName := a.nsResolver.ResolveWithType(req)
	if len(endpoint.Name) > 0 {
		resName = resName + "." + endpoint.Name
	}

	// determine the resource path. there is an inconsistency in how resolvers
	// use method, some use it as Users.ReadUser (the rpc method), and others
	// use it as the HTTP method, e.g GET. TODO: Refactor this to make it consistent.
	resEndpoint := endpoint.Path
	if len(endpoint.Path) == 0 {
		resEndpoint = endpoint.Method
	}

	// Perform the verification check to see if the account has access to
	// the resource they're requesting
	res := &auth.Resource{Type: "service", Name: resName, Endpoint: resEndpoint}
	if err := a.auth.Verify(acc, res, auth.VerifyContext(req.Context())); err == nil {
		// The account has the necessary permissions to access the resource
		a.handler.ServeHTTP(w, req)
		return
	}

	// The account is set, but they don't have enough permissions, hence
	// we return a forbidden error.
	if acc != nil {
		http.Error(w, "Forbidden request", 403)
		return
	}

	// If there is no auth login url set, 401
	loginURL := a.auth.Options().LoginURL
	if loginURL == "" {
		http.Error(w, "unauthorized request", 401)
		return
	}

	// Redirect to the login path
	params := url.Values{"redirect_to": {req.URL.String()}}
	loginWithRedirect := fmt.Sprintf("%v?%v", loginURL, params.Encode())
	http.Redirect(w, req, loginWithRedirect, http.StatusTemporaryRedirect)
}
