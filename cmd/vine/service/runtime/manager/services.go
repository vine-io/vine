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

package manager

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/namespace"
)

// service is the object persisted in the store
type service struct {
	Service *runtime.Service       `json:"service"`
	Options *runtime.CreateOptions `json:"options"`
}

const (
	// servicePrefix is prefixed to the key for service records
	servicePrefix = "service:"
)

// key to write the service to the store under, e.g:
// "service/foo/go.vine.service.bar:latest"
func (s *service) Key() string {
	return servicePrefix + s.Options.Namespace + ":" + s.Service.Name + ":" + s.Service.Version
}

// createService writes the service to the store
func (m *manager) createService(svc *runtime.Service, opts *runtime.CreateOptions) error {
	s := &service{svc, opts}

	bytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return m.options.Store.Write(&store.Record{Key: s.Key(), Value: bytes})
}

// readServices returns all the services in a given namespace. If a service name and
// version are provided it will filter using these as well
func (m *manager) readServices(namespace string, svc *runtime.Service) ([]*runtime.Service, error) {
	prefix := servicePrefix + namespace + ":"
	if len(svc.Name) > 0 {
		prefix += svc.Name + ":"
	}
	if len(svc.Name) > 0 && len(svc.Version) > 0 {
		prefix += svc.Version
	}

	recs, err := m.options.Store.Read(prefix, store.ReadPrefix())
	if err != nil {
		return nil, err
	} else if len(recs) == 0 {
		return make([]*runtime.Service, 0), nil
	}

	svcs := make([]*runtime.Service, 0, len(recs))
	for _, r := range recs {
		var s *service
		if err := json.Unmarshal(r.Value, &s); err != nil {
			return nil, err
		}
		svcs = append(svcs, s.Service)
	}

	return svcs, nil
}

// deleteSevice from the store
func (m *manager) deleteService(namespace string, svc *runtime.Service) error {
	obj := &service{svc, &runtime.CreateOptions{Namespace: namespace}}
	return m.options.Store.Delete(obj.Key())
}

// listNamespaces of the services in the store
func (m *manager) listNamespaces() ([]string, error) {
	recs, err := m.options.Store.Read(servicePrefix, store.ReadPrefix())
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return []string{namespace.DefaultNamespace}, nil
	}

	namespaces := make([]string, 0, len(recs))
	for _, rec := range recs {
		// key is formatted 'prefix:namespace:name:version'
		if comps := strings.Split(rec.Key, ":"); len(comps) == 4 {
			namespaces = append(namespaces, comps[1])
		} else {
			return nil, fmt.Errorf("Invalid key: %v", rec.Key)
		}
	}

	return unique(namespaces), nil
}

// unique is a helper method to filter a slice of strings
// down to unique entries
func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
