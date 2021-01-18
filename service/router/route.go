// Copyright 2020 lack
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

import "hash/fnv"

var (
	// DefaultLink is default network link
	DefaultLink = "local"
	// DefaultLocalMetric is default route cost for a local route
	DefaultLocalMetric int64 = 1
)

// Route is network route
type Route struct {
	// Service is destination service name
	Service string
	// Address is service node address
	Address string
	// Gateway is route gateway
	Gateway string
	// Network is network address
	Network string
	// Router is router id
	Router string
	// Link is network link
	Link string
	// Metric is the route cost metric
	Metric int64
}

// Hash returns route hash sum
func (r *Route) Hash() uint64 {
	h := fnv.New64()
	h.Reset()
	h.Write([]byte(r.Service + r.Address + r.Gateway + r.Network + r.Router + r.Link))
	return h.Sum64()
}
