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

// Package provider lets you abstract away any number of DNS providers
package provider

import (
	"github.com/lack-io/vine/proto/services/network/dns"
)

// Provider is an interface for interacting with a DNS provider
type Provider interface {
	// Advertise creates records in DNS
	Advertise(...*dns.Record) error
	// Remove removes records from DNS
	Remove(...*dns.Record) error
	// Resolve looks up a record in DNS
	Resolve(name, recordType string) ([]*dns.Record, error)
}
