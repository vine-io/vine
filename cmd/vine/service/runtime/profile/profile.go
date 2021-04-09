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

// Package profile is for specific profiles
// @todo this package is the definition of cruft and
// should be rewritten in a more elegant way
package profile

// Local is a profile for local environments
func Local() []string {
	return []string{}
}

// Server is a profile for running things through vine server
// eg runtime config etc will use actual services.
func Server() []string {
	return []string{
		"VINE_AUTH=service",
		"VINE_BROKER=service",
		"VINE_REGISTRY=service",
		"VINE_ROUTER=service",
		"VINE_RUNTIME=service",
		"VINE_STORE=service",
		"VINE_CONFIG=service",
	}
}

func ServerCLI() []string {
	return []string{
		"VINE_AUTH=service",
		"VINE_BROKER=service",
		"VINE_REGISTRY=service",
		"VINE_ROUTER=service",
		"VINE_RUNTIME=service",
		"VINE_STORE=service",
		"VINE_CONFIG=service",
	}
}

// Kubernetes is a profile for kubernetes
func Kubernetes() []string {
	return []string{
		"VINE_AUTH=service",
		"VINE_BROKER=service",
		"VINE_CONFIG=service",
		"VINE_NETWORK=service",
		"VINE_REGISTRY=service",
		"VINE_ROUTER=service",
		"VINE_RUNTIME=service",
		"VINE_STORE=service",
		"VINE_AUTH_ADDRESS=vine-auth:8010",
		"VINE_BROKER_ADDRESS=vine-store:8001",
		"VINE_NETWORK_ADDRESS=vine-network:8080",
		"VINE_REGISTRY_ADDRESS=vine-registry:8000",
		"VINE_ROUTER_ADDRESS=vine-runtime:8084",
		"VINE_RUNTIME_ADDRESS=vine-runtime:8088",
		"VINE_STORE_ADDRESS=vine-store:8002",
	}
}

// Platform is a platform profile
func Platform() []string {
	return []string{
		// TODO: debug, monitor, etc
		"VINE_AUTH=service",
		"VINE_BROKER=service",
		"VINE_CONFIG=service",
		"VINE_NETWORK=service",
		"VINE_REGISTRY=service",
		"VINE_ROUTER=service",
		"VINE_RUNTIME=service",
		"VINE_STORE=service",
		// now set the addresses
		"VINE_AUTH_ADDRESS=vine-auth.default.svc:8010",
		"VINE_BROKER_ADDRESS=vine-store.default.svc:8001",
		"VINE_NETWORK_ADDRESS=vine-network.default.svc:8080",
		"VINE_REGISTRY_ADDRESS=vine-registry.default.svc:8000",
		"VINE_ROUTER_ADDRESS=vine-runtime.default.svc:8084",
		"VINE_RUNTIME_ADDRESS=vine-runtime.default.svc:8088",
		"VINE_STORE_ADDRESS=vine-store.default.svc:8002",
		// set the athens proxy to speedup builds
		"GOPROXY=http://athens-proxy",
	}
}

// Platform is a platform profile
func PlatformCLI() []string {
	return []string{
		// TODO: debug, monitor, etc
		"VINE_AUTH=service",
		"VINE_BROKER=service",
		"VINE_CONFIG=service",
		"VINE_REGISTRY=service",
		"VINE_ROUTER=service",
		"VINE_RUNTIME=service",
		"VINE_STORE=service",
	}
}
