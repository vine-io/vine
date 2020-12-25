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

package handler

import (
	"net/http"

	"github.com/lack-io/vine/proto/errors"
	"github.com/lack-io/vine/service"
	"github.com/lack-io/vine/service/api/handler"
	"github.com/lack-io/vine/service/api/handler/event"
	"github.com/lack-io/vine/service/api/router"
	"github.com/lack-io/vine/service/client"

	// TODO: only import handler package
	aapi "github.com/lack-io/vine/service/api/handler/api"
	ahttp "github.com/lack-io/vine/service/api/handler/http"
	arpc "github.com/lack-io/vine/service/api/handler/rpc"
	aweb "github.com/lack-io/vine/service/api/handler/web"
)

type metaHandler struct {
	c  client.Client
	r  router.Router
	ns func(*http.Request) string
}

func (m *metaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, err := m.r.Route(r)
	if err != nil {
		er := errors.InternalServerError(m.ns(r), err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}

	// TODO: don't do this ffs
	switch service.Endpoint.Handler {
	// web socket handler
	case aweb.Handler:
		aweb.WithService(service, handler.WithClient(m.c)).ServeHTTP(w, r)
	// proxy handler
	case "proxy", ahttp.Handler:
		ahttp.WithService(service, handler.WithClient(m.c)).ServeHTTP(w, r)
	// rpcx handler
	case arpc.Handler:
		arpc.WithService(service, handler.WithClient(m.c)).ServeHTTP(w, r)
	// event handler
	case event.Handler:
		ev := event.NewHandler(
			handler.WithNamespace(m.ns(r)),
			handler.WithClient(m.c),
		)
		ev.ServeHTTP(w, r)
	// api handler
	case aapi.Handler:
		aapi.WithService(service, handler.WithClient(m.c)).ServeHTTP(w, r)
	// default handler: rpc
	default:
		arpc.WithService(service, handler.WithClient(m.c)).ServeHTTP(w, r)
	}
}

// Meta is a http.Handler that routes based on endpoint metadata
func Meta(s service.Service, r router.Router, ns func(*http.Request) string) http.Handler {
	return &metaHandler{
		c:  s.Client(),
		r:  r,
		ns: ns,
	}
}
