// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import "github.com/lack-io/cli"

// Flags common to all clients
var Flags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "local",
		Usage: "Enable local only development: Defaults to true.",
	},
	&cli.BoolFlag{
		Name:    "enable-acme",
		Usage:   "Enables ACME support via Let's Encrypt. ACME hosts should also be specified.",
		EnvVars: []string{"VINE_ENABLE_ACME"},
	},
	&cli.StringFlag{
		Name:    "acme-hosts",
		Usage:   "Comma separated list of hostnames to manage ACME certs for",
		EnvVars: []string{"VINE_ACME_HOSTS"},
	},
	&cli.StringFlag{
		Name:    "acme-provider",
		Usage:   "The provider that will be used to communicate with Let's Encrypt. Valid options: autocert, certmagic",
		EnvVars: []string{"VINE_ACME_PROVIDER"},
	},
	&cli.BoolFlag{
		Name:    "enable-tls",
		Usage:   "Enable TLS support. Expects cert and key file to be specified",
		EnvVars: []string{"VINE_ENABLE_TLS"},
	},
	&cli.StringFlag{
		Name:    "tls-cert-file",
		Usage:   "Path to the TLS Certificate file",
		EnvVars: []string{"VINE_TLS_CERT_FILE"},
	},
	&cli.StringFlag{
		Name:    "tls-key-file",
		Usage:   "Path to the TLS Key file",
		EnvVars: []string{"VINE_TLS_KEY_FILE"},
	},
	&cli.StringFlag{
		Name:    "tls-client-ca-file",
		Usage:   "Path to the TLS CA file to verify clients against",
		EnvVars: []string{"VINE_TLS_CLIENT_CA_FILE"},
	},
}
