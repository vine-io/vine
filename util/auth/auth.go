// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"fmt"
	"time"

	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/logger"
)

// Generate generates a service account for and continually
// refreshes the access token.
func Generate(id string, name string, a auth.Auth) error {
	// extract the account creds from options, these can be set by flags
	accID := a.Options().ID
	accSecret := a.Options().Secret

	// if no credentials were provided, generate an account
	if len(accID) == 0 || len(accSecret) == 0 {
		name := fmt.Sprintf("%v-%v", name, id)

		opts := []auth.GenerateOption{
			auth.WithType("service"),
			auth.WithScopes("service"),
		}

		acc, err := a.Generate(name, opts...)
		if err != nil {
			return err
		}
		logger.Debugf("Auth [%v] Authenticated as %v issued by %v", a, name, acc.Issuer)

		accID = acc.ID
		accSecret = acc.Secret
	}

	// generate the first token
	token, err := a.Token(
		auth.WithCredentials(accID, accSecret),
		auth.WithExpiry(time.Minute*10),
	)
	if err != nil {
		return err
	}

	// set the credentials and token in auth options
	a.Init(
		auth.ClientToken(token),
		auth.Credentials(accID, accSecret),
	)

	// periodically check to see if the token needs refreshing
	go func() {
		timer := time.NewTicker(time.Second * 15)

		for {
			<-timer.C

			// don't refresh the token if it's not close to expiring
			tok := a.Options().Token
			if tok.Expiry.Unix() > time.Now().Add(time.Minute).Unix() {
				continue
			}

			// generate the first token
			tok, err := a.Token(
				auth.WithToken(tok.RefreshToken),
				auth.WithExpiry(time.Minute*10),
			)
			if err != nil {
				logger.Warnf("[Auth] Error refreshing token: %v", err)
				continue
			}

			// set the token
			a.Init(auth.ClientToken(tok))
		}
	}()

	return nil
}

// TokenCookieName is the name of the cookie which stores the auth token
const TokenCookieName = "vine-token"

// SystemRules are the default rules which are applied to the runtime services
var SystemRules = []*auth.Rule{
	&auth.Rule{
		ID:       "default",
		Scope:    "*",
		Resource: &auth.Resource{Type: "*", Name: "*", Endpoint: "*"},
	},
	&auth.Rule{
		ID:       "auth-public",
		Scope:    "",
		Resource: &auth.Resource{Type: "service", Name: "go.vine.auth", Endpoint: "*"},
	},
	&auth.Rule{
		ID:       "registry-get",
		Scope:    "",
		Resource: &auth.Resource{Type: "service", Name: "go.vine.registry", Endpoint: "Registry.GetService"},
	},
	&auth.Rule{
		ID:       "registry-list",
		Scope:    "",
		Resource: &auth.Resource{Type: "service", Name: "go.vine.registry", Endpoint: "Registry.ListServices"},
	},
}
