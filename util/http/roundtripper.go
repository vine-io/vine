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

package http

import (
	"errors"
	"net/http"

	"github.com/lack-io/vine/service/client/selector"
)

type roundTripper struct {
	rt   http.RoundTripper
	st   selector.Strategy
	opts Options
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s, err := r.opts.Registry.GetService(req.URL.Host)
	if err != nil {
		return nil, err
	}

	next := r.st(s)

	// rudimentary retry 3 times
	for i := 0; i < 3; i++ {
		n, err := next()
		if err != nil {
			continue
		}
		req.URL.Host = n.Address
		w, err := r.rt.RoundTrip(req)
		if err != nil {
			continue
		}
		return w, nil
	}

	return nil, errors.New("failed request")
}
