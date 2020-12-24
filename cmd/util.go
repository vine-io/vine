// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/client/cli/namespace"
	clitoken "github.com/lack-io/vine/client/cli/token"
	"github.com/lack-io/vine/client/cli/util"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/logger"
)

func formatErr(err error) string {
	switch v := err.(type) {
	case *errors.Error:
		return upcaseInitial(v.Detail)
	default:
		return upcaseInitial(err.Error())
	}
}

func upcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// setupAuthForCLI handles exchanging refresh tokens to access tokens
// The structure of the local vine userconfig file is the following:
// vine.auth.[envName].token: temporary access token
// vine.auth.[envName].refresh-token: long lived refresh token
// vine.auth.[envName].expiry: expiration time of the access token, seconds since Unix epoch.
func setupAuthForCLI(ctx *cli.Context) error {
	env, err := util.GetEnv(ctx)
	if err != nil {
		return err
	}
	ns, err := namespace.Get(env.Name)
	if err != nil {
		return err
	}

	tok, err := clitoken.Get(ctx)
	if err != nil {
		return err
	}

	// If there is no refresh token, do not try to refresh it
	if len(tok.RefreshToken) == 0 {
		return nil
	}

	// Check if token is valid
	if time.Now().Before(tok.Expiry.Add(time.Minute * -1)) {
		auth.DefaultAuth.Init(
			auth.ClientToken(tok),
			auth.Issuer(ns),
		)
		return nil
	}

	// Get new access token from refresh token if it's close to expiry
	tok, err = auth.Token(
		auth.WithToken(tok.RefreshToken),
		auth.WithTokenIssuer(ns),
		auth.WithExpiry(time.Minute*10),
	)
	if err != nil {
		return nil
	}

	// Save the token to user config file
	auth.DefaultAuth.Init(
		auth.ClientToken(tok),
		auth.Issuer(ns),
	)
	return clitoken.Save(ctx, tok)
}

// setupAuthForService generates auth credentials for the service
func setupAuthForService() error {
	opts := auth.DefaultAuth.Options()

	// extract the account creds from options, these can be set by flags
	accID := opts.ID
	accSecret := opts.Secret

	// if no credentials were provided, self generate an account
	if len(accID) == 0 || len(accSecret) == 0 {
		opts := []auth.GenerateOption{
			auth.WithType("service"),
			auth.WithScopes("service"),
		}

		acc, err := auth.Generate(uuid.New().String(), opts...)
		if err != nil {
			return err
		}
		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("Auth [%v] Generated an auth account", auth.DefaultAuth.String())
		}

		accID = acc.ID
		accSecret = acc.Secret
	}

	// generate the first token
	token, err := auth.Token(
		auth.WithCredentials(accID, accSecret),
		auth.WithExpiry(time.Minute*10),
	)
	if err != nil {
		return err
	}

	// set the credentials and token in auth options
	auth.DefaultAuth.Init(
		auth.ClientToken(token),
		auth.Credentials(accID, accSecret),
	)
	return nil
}

// refreshAuthToken if it is close to expiring
func refreshAuthToken() {
	// can't refresh a token we don't have
	if auth.DefaultAuth.Options().Token == nil {
		return
	}

	t := time.NewTicker(time.Second * 15)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			// don't refresh the token if it's not close to expiring
			tok := auth.DefaultAuth.Options().Token
			if tok.Expiry.Unix() > time.Now().Add(time.Minute).Unix() {
				continue
			}

			// generate the first token
			tok, err := auth.Token(
				auth.WithToken(tok.RefreshToken),
				auth.WithExpiry(time.Minute*10),
			)
			if err == auth.ErrInvalidToken {
				logger.Warnf("[Auth] Refresh token expired, regenerating using account credentials")

				tok, err = auth.Token(
					auth.WithCredentials(
						auth.DefaultAuth.Options().ID,
						auth.DefaultAuth.Options().Secret,
					),
					auth.WithExpiry(time.Minute*10),
				)
			} else if err != nil {
				logger.Warnf("[Auth] Error refreshing token: %v", err)
				continue
			}

			// set the token
			logger.Debugf("Auth token refreshed, expires at %v", tok.Expiry.Format(time.UnixDate))
			auth.DefaultAuth.Init(auth.ClientToken(tok))
		}
	}
}
