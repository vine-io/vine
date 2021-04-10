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

	"github.com/gofiber/fiber/v2"
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
	return func() fiber.Handler {
		//return authWrapper{
		//	handler:    h,
		//	resolver:   r,
		//	nsResolver: nr,
		//	auth:       auth.DefaultAuth,
		//}
		return nil
	}
}

type authWrapper struct {
	handler    func(c *fiber.Ctx) error
	auth       auth.Auth
	resolver   resolver.Resolver
	nsResolver *namespace.Resolver
}

func (a authWrapper) Handle(c *fiber.Ctx) error {
	// Determine the namespace and set it in the header
	ns := c.Get(namespace.NamespaceKey)
	if len(ns) == 0 {
		ns = a.nsResolver.Resolve(c)
		c.Set(namespace.NamespaceKey, ns)
	}

	// Set the metadata so we can access it in vine api / web
	r := ctx.NewRequestCtx(c, ctx.FromRequest(c))

	// Extract the token from the request
	var token string
	if header := r.Get("Authorization"); len(header) > 0 {
		// Extract the auth token from the request
		if strings.HasPrefix(header, auth.BearerScheme) {
			token = header[len(auth.BearerScheme):]
		}
	} else {
		// Get the token out the cookies if not provided in headers
		if cookie := r.Cookies("vine-token"); cookie != "" {
			token = strings.TrimPrefix(cookie, inauth.TokenCookieName+"=")
			c.Set("Authorization", auth.BearerScheme+token)
		}
	}

	// Get the account using the token, some are unauthenticated, so the lack of an
	// account doesn't necesserially mean a forbidden request
	acc, _ := a.auth.Inspect(token)

	// Ensure the accounts issuer matches the namespace being requested
	if acc != nil && len(acc.Issuer) > 0 && acc.Issuer != ns {
		return fiber.NewError(403, "Account not issued by "+ns)
	}

	// Determine the name of the service being requested
	endpoint, err := a.resolver.Resolve(r.Ctx)
	if err == resolver.ErrInvalidPath || err == resolver.ErrNotFound {
		// a file not served by the resolver has been requested (e.g. favicon.ico)
		endpoint = &resolver.Endpoint{Path: r.Path()}
	} else if err != nil {
		log.Error(err)
		return fiber.NewError(500, err.Error())
	} else {
		// set the endpoint in the context so it can be used to resolve
		// the request later
		cx := context.WithValue(r.Context(), resolver.Endpoint{}, endpoint)
		r = r.Clone(cx)
	}

	// construct the resource name, e.g. home => go.vine.web.home
	resName := a.nsResolver.ResolveWithType(c)
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
	if err := a.auth.Verify(acc, res, auth.VerifyContext(r.Context())); err == nil {
		// The account has the necessary permissions to access the resource
		return a.handler(r.Ctx)
	}

	// The account is set, but they don't have enough permissions, hence
	// we return a forbidden error.
	if acc != nil {
		return fiber.NewError(403, "Forbidden request")
	}

	// If there is no auth login url set, 401
	loginURL := a.auth.Options().LoginURL
	if loginURL == "" {
		return fiber.NewError(401, "unauthorized request")
	}

	// Redirect to the login path
	params := url.Values{"redirect_to": {c.Request().URI().String()}}
	loginWithRedirect := fmt.Sprintf("%v?%v", loginURL, params.Encode())
	return c.Redirect(loginWithRedirect, http.StatusTemporaryRedirect)
}
