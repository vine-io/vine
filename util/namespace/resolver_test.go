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
