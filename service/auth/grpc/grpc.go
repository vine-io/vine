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
package grpc

import (
	"context"
	"strings"
	"time"

	pb "github.com/lack-io/vine/proto/auth"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/auth/rules"
	"github.com/lack-io/vine/service/auth/token"
	"github.com/lack-io/vine/service/auth/token/jwt"
	"github.com/lack-io/vine/service/client"
)

// gRPC is the service implementation of the Auth interface
type gRPC struct {
	options auth.Options
	auth    pb.AuthService
	rules   pb.RulesService
	jwt     token.Provider
}

func (s *gRPC) String() string {
	return "grpc"
}

func (s *gRPC) Init(opts ...auth.Option) {
	for _, o := range opts {
		o(&s.options)
	}

	if s.options.Client == nil {
		s.options.Client = client.DefaultClient
	}

	s.auth = pb.NewAuthService("go.vine.auth", s.options.Client)
	s.rules = pb.NewRulesService("go.vine.auth", s.options.Client)

	// if we have a JWT public key passed as an option,
	// we can decode tokens with the type "JWT" locally
	// and not have to make an RPC call
	if key := s.options.PublicKey; len(key) > 0 {
		s.jwt = jwt.NewTokenProvider(token.WithPublicKey(key))
	}
}

func (s *gRPC) Options() auth.Options {
	return s.options
}

// Generate a new account
func (s *gRPC) Generate(id string, opts ...auth.GenerateOption) (*auth.Account, error) {
	options := auth.NewGenerateOptions(opts...)

	rsp, err := s.auth.Generate(context.TODO(), &pb.GenerateRequest{
		Id:       id,
		Type:     options.Type,
		Secret:   options.Secret,
		Scopes:   options.Scopes,
		Metadata: options.Metadata,
		Provider: options.Provider,
	})
	if err != nil {
		return nil, err
	}

	return serializeAccount(rsp.Account), nil
}

// Grant access to a resource
func (s *gRPC) Grant(rule *auth.Rule) error {
	access := pb.Access_UNKNOWN
	if rule.Access == auth.AccessGranted {
		access = pb.Access_GRANTED
	} else if rule.Access == auth.AccessDenied {
		access = pb.Access_DENIED
	}

	_, err := s.rules.Create(context.TODO(), &pb.CreateRequest{
		Rule: &pb.Rule{
			Id:       rule.ID,
			Scope:    rule.Scope,
			Priority: rule.Priority,
			Access:   access,
			Resource: &pb.Resource{
				Type:     rule.Resource.Type,
				Name:     rule.Resource.Name,
				Endpoint: rule.Resource.Endpoint,
			},
		},
	})

	return err
}

// Revoke access to a resource
func (s *gRPC) Revoke(rule *auth.Rule) error {
	_, err := s.rules.Delete(context.TODO(), &pb.DeleteRequest{
		Id: rule.ID,
	})

	return err
}

func (s *gRPC) Rules(opts ...auth.RulesOption) ([]*auth.Rule, error) {
	var options auth.RulesOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	rsp, err := s.rules.List(options.Context, &pb.ListRequest{}, client.WithCache(time.Second*30))
	if err != nil {
		return nil, err
	}

	rules := make([]*auth.Rule, len(rsp.Rules))
	for i, r := range rsp.Rules {
		rules[i] = serializeRule(r)
	}

	return rules, nil
}

// Verify an account has access to a resource
func (s *gRPC) Verify(acc *auth.Account, res *auth.Resource, opts ...auth.VerifyOption) error {
	var options auth.VerifyOptions
	for _, o := range opts {
		o(&options)
	}

	rs, err := s.Rules(auth.RulesContext(options.Context))
	if err != nil {
		return err
	}

	return rules.Verify(rs, acc, res)
}

// Inspect a token
func (s *gRPC) Inspect(token string) (*auth.Account, error) {
	// try to decode JWT locally and fall back to srv if an error occurs
	if len(strings.Split(token, ".")) == 3 && s.jwt != nil {
		return s.jwt.Inspect(token)
	}

	// the token is not a JWT or we do not have the keys to decode it,
	// fall back to the auth service
	rsp, err := s.auth.Inspect(context.TODO(), &pb.InspectRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return serializeAccount(rsp.Account), nil
}

// Token generation using an account ID and secret
func (s *gRPC) Token(opts ...auth.TokenOption) (*auth.Token, error) {
	options := auth.NewTokenOptions(opts...)

	rsp, err := s.auth.Token(context.Background(), &pb.TokenRequest{
		Id:           options.ID,
		Secret:       options.Secret,
		RefreshToken: options.RefreshToken,
		TokenExpiry:  int64(options.Expiry.Seconds()),
	})
	if err != nil {
		return nil, err
	}

	return serializeToken(rsp.Token), nil
}

func serializeToken(t *pb.Token) *auth.Token {
	return &auth.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Created:      time.Unix(t.Created, 0),
		Expiry:       time.Unix(t.Expiry, 0),
	}
}

func serializeAccount(a *pb.Account) *auth.Account {
	return &auth.Account{
		ID:       a.Id,
		Secret:   a.Secret,
		Issuer:   a.Issuer,
		Metadata: a.Metadata,
		Scopes:   a.Scopes,
	}
}

func serializeRule(r *pb.Rule) *auth.Rule {
	var access auth.Access
	if r.Access == pb.Access_GRANTED {
		access = auth.AccessGranted
	} else {
		access = auth.AccessDenied
	}

	return &auth.Rule{
		ID:       r.Id,
		Scope:    r.Scope,
		Access:   access,
		Priority: r.Priority,
		Resource: &auth.Resource{
			Type:     r.Resource.Type,
			Name:     r.Resource.Name,
			Endpoint: r.Resource.Endpoint,
		},
	}
}

// NewAuth returns a new instance of the Auth service
func NewAuth(opts ...auth.Option) auth.Auth {
	options := auth.NewOptions(opts...)
	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	return &gRPC{
		auth:    pb.NewAuthService("go.vine.auth", options.Client),
		rules:   pb.NewRulesService("go.vine.auth", options.Client),
		options: options,
	}
}
