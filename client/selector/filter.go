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

package selector

import (
	"github.com/lack-io/vine/registry"
)

// FilterEndpoint is an endpoint based Select Filter which will
// only return services with the endpoint specified.
func FilterEndpoint(name string) Filter {
	return func(old []*registry.Service) []*registry.Service {
		var services []*registry.Service

		for _, service := range old {
			for _, ep := range service.Endpoints {
				if ep.Name == name {
					services = append(services, service)
					break
				}
			}
		}

		return services
	}
}

// FilterLabel is a label based Select Filter which will
// only return services with the label specified.
func FilterLabel(key, val string) Filter {
	return func(old []*registry.Service) []*registry.Service {
		var services []*registry.Service

		for _, service := range old {
			serv := new(registry.Service)
			var nodes []*registry.Node

			for _, node := range service.Nodes {
				if node.Metadata == nil {
					continue
				}

				if node.Metadata[key] == val {
					nodes = append(nodes, node)
				}
			}

			// only add service if there's some nodes
			if len(nodes) > 0 {
				// copy
				*serv = *service
				serv.Nodes = nodes
				services = append(services, serv)
			}
		}

		return services
	}
}

// FilterVersion is a version based Select Filter which will
// only return services with the version specified.
func FilterVersion(version string) Filter {
	return func(old []*registry.Service) []*registry.Service {
		var services []*registry.Service

		for _, service := range old {
			if service.Version == version {
				services = append(services, service)
			}
		}

		return services
	}
}
