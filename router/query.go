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

package router

// QueryOption sets routing table query options
type QueryOption func(*QueryOptions)

// QueryOptions are routing table query options
type QueryOptions struct {
	// Service is destination service name
	Service string
	// Address of the service
	Address string
	// Gateway is route gateway
	Gateway string
	// Network is network address
	Network string
	// Router is router id
	Router string
	// Strategy is routing strategy
	Strategy Strategy
}

// QueryService sets service to query
func QueryService(s string) QueryOption {
	return func(o *QueryOptions) {
		o.Service = s
	}
}

// QueryAddress sets service to query
func QueryAddress(a string) QueryOption {
	return func(o *QueryOptions) {
		o.Address = a
	}
}

// QueryGateway sets network name to query
func QueryGateway(n string) QueryOption {
	return func(o *QueryOptions) {
		o.Gateway = n
	}
}

// QueryNetwork sets network name to query
func QueryNetwork(n string) QueryOption {
	return func(o *QueryOptions) {
		o.Network = n
	}
}

// QueryRouter sets router id to query
func QueryRouter(r string) QueryOption {
	return func(o *QueryOptions) {
		o.Router = r
	}
}

// QueryStrategy sets strategy to query
func QueryStrategy(s Strategy) QueryOption {
	return func(o *QueryOptions) {
		o.Strategy = s
	}
}

// NewQuery creates new query and returns it
func NewQuery(opts ...QueryOption) QueryOptions {
	// default options
	qopts := QueryOptions{
		Service:  "*",
		Address:  "*",
		Gateway:  "*",
		Network:  "*",
		Router:   "*",
		Strategy: AdvertiseAll,
	}

	for _, o := range opts {
		o(&qopts)
	}

	return qopts
}
