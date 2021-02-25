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

package api

import (
	"path"
	"regexp"
	"strings"
)

var (
	proxyRe   = regexp.MustCompile("^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$")
	versionRe = regexp.MustCompilePOSIX("^v[0-9]+$")
)

// Translates /foo/bar/zool into api service go.vine.api.foo method Bar.Zool
// Translates /foo/bar into api service go.vine.api.foo method Foo.Bar
func apiRoute(p string) (string, string) {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	parts := strings.Split(p, "/")

	// if we have 1 part assume name Name.Call
	if len(parts) == 1 && len(parts[0]) > 0 {
		return parts[0], methodName(append(parts, "Call"))
	}

	// If we've got two or less parts
	// Use first part as service
	// Use all parts as method
	if len(parts) <= 2 {
		name := parts[0]
		return name, methodName(parts)
	}

	// Treat /v[0-9]+ as versioning where we have 3 parts
	// /v1/foo/bar => service: v1.foo method: Foo.bar
	if len(parts) == 3 && versionRe.Match([]byte(parts[0])) {
		name := strings.Join(parts[:len(parts)-1], ".")
		return name, methodName(parts[len(parts)-2:])
	}

	// Service is everything minus last two parts
	// Method is the last two parts
	name := strings.Join(parts[:len(parts)-2], ".")
	return name, methodName(parts[len(parts)-2:])
}

func proxyRoute(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) < 2 {
		return ""
	}

	var service string
	var alias string

	// /[service]/methods
	if len(parts) > 2 {
		// /v1/[service]
		if versionRe.MatchString(parts[1]) {
			service = parts[1] + "." + parts[2]
			alias = parts[2]
		} else {
			service = parts[1]
			alias = parts[1]
		}
		// /[service]
	} else {
		service = parts[1]
		alias = parts[1]
	}

	// check service name is valid
	if !proxyRe.MatchString(alias) {
		return ""
	}

	return service
}

func methodName(parts []string) string {
	for i, part := range parts {
		parts[i] = toCamel(part)
	}

	return strings.Join(parts, ".")
}

func toCamel(s string) string {
	words := strings.Split(s, "-")
	var out string
	for _, word := range words {
		out += strings.Title(word)
	}
	return out
}
