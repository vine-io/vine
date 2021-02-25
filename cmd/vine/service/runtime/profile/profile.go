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
