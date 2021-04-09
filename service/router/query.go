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
