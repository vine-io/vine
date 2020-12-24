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

package namespace

import (
	"net/http"
	"net/url"
	"testing"
)

func TestResolveWithType(t *testing.T) {
	tt := []struct {
		Name        string
		Namespace   string
		ServiceType string
		Host        string
		Result      string
	}{
		{
			Name:        "A fixed namespace with web type",
			ServiceType: "web",
			Namespace:   "foobar",
			Host:        "vine.mu",
			Result:      "foobar.web",
		},
		{
			Name:        "A dynamic namespace with a type web and the vine.mu domain",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "vine.mu",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type api and the vine.mu domain",
			ServiceType: "api",
			Namespace:   "domain",
			Host:        "vine.mu",
			Result:      DefaultNamespace + ".api",
		},
		{
			Name:        "A dynamic namespace with a type web and a effective top level domain",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "vine.com.au",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web and the web.vine.mu subdomain",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "web.vine.mu",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web and a vine.mu subdomain",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "foo.vine.mu",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web and top level domain host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "myapp.com",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web subdomain host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "staging.myapp.com",
			Result:      "staging.web",
		},
		{
			Name:        "A dynamic namespace with a type web and multi-level subdomain host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "staging.myapp.m3o.app",
			Result:      "myapp.staging.web",
		},
		{
			Name:        "A dynamic namespace with a type web and dev host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "127.0.0.1",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web and localhost host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "localhost",
			Result:      DefaultNamespace + ".web",
		},
		{
			Name:        "A dynamic namespace with a type web and IP host",
			ServiceType: "web",
			Namespace:   "domain",
			Host:        "81.151.101.146",
			Result:      DefaultNamespace + ".web",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := NewResolver(tc.ServiceType, tc.Namespace)
			result := r.ResolveWithType(&http.Request{URL: &url.URL{Host: tc.Host}})
			if result != tc.Result {
				t.Errorf("Expected namespace %v for host %v with service type %v and namespace %v, actually got %v", tc.Result, tc.Host, tc.ServiceType, tc.Namespace, result)
			}
		})
	}
}
