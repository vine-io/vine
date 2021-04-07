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

package context

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/lack-io/vine/util/context/metadata"
)

func FromRequest(r *http.Request) context.Context {
	ctx := r.Context()
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	for k, v := range r.Header {
		md[textproto.CanonicalMIMEHeaderKey(k)] = strings.Join(v, ",")
	}
	if v, ok := md.Get("X-Forwarded-For"); ok {
		md["X-Forwarded-For"] = v + "," + r.RemoteAddr
	} else {
		md["X-Forwarded-For"] = v
	}
	md["Host"] = r.Host
	// pass http method
	md["Method"] = r.Method
	return metadata.NewContext(ctx, md)
}
