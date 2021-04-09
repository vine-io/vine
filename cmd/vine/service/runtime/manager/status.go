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
	"time"

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store"
)

// statusPrefix is prefixed to every status key written to the memory store
const statusPrefix = "status:"

// serviceStatus contains the runtime specific information for a service
type serviceStatus struct {
	Status string
	Error  string
}

// statusPollFrequency is the max frequency the manager will check for new statuses in the runtime
var statusPollFrequency = time.Second * 30

// watchStatus calls syncStatus periodically and should be run in a seperate go routine
func (m *manager) watchStatus() {
	ticker := time.NewTicker(statusPollFrequency)

	for {
		m.syncStatus()
		<-ticker.C
	}
}

// syncStatus calls the managed runtime, gets the serviceStatus for all services listed in the
// store and writes it to the memory store
func (m *manager) syncStatus() {
	namespaces, err := m.listNamespaces()
	if err != nil {
		log.Warnf("Error listing namespaces: %v", err)
		return
	}

	for _, ns := range namespaces {
		svcs, err := m.Runtime.Read(runtime.ReadNamespace(ns))
		if err != nil {
			log.Warnf("Error reading namespace %v: %v", ns, err)
			return
		}

		for _, svc := range svcs {
			if err := m.cacheStatus(ns, svc); err != nil {
				log.Warnf("Error caching status: %v", err)
				return
			}
		}
	}
}

// cacheStatus writes a services status to the memory store which is then later returned in service
// metadata on Runtime.Read
func (m *manager) cacheStatus(ns string, svc *runtime.Service) error {
	// errors / status is returned from the underlying runtime using svc.Metadata. TODO: Consider
	// changing this so status / error are attributes on runtime.Service.
	if svc.Metadata == nil {
		return fmt.Errorf("Service %v:%v (%v) is missing metadata", svc.Name, svc.Version, ns)
	}

	key := fmt.Sprintf("%v%v:%v:%v", statusPrefix, ns, svc.Name, svc.Version)
	val := &serviceStatus{Status: svc.Metadata["status"], Error: svc.Metadata["error"]}

	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return m.cache.Write(&store.Record{Key: key, Value: bytes})
}

// listStautuses returns all the statuses for the services in a given namespace with 'name:version'
// as the format used for the keys in the map.
func (m *manager) listStatuses(ns string) (map[string]*serviceStatus, error) {
	recs, err := m.cache.Read(statusPrefix+ns+":", store.ReadPrefix())
	if err != nil {
		return nil, fmt.Errorf("Error listing statuses from the store for namespace %v: %v", ns, err)
	}

	statuses := make(map[string]*serviceStatus, len(recs))

	for _, rec := range recs {
		var status *serviceStatus
		if err := json.Unmarshal(rec.Value, &status); err != nil {
			return nil, err
		}

		// record keys are formatted: 'prefix:namespace:name:version'
		if comps := strings.Split(rec.Key, ":"); len(comps) == 4 {
			statuses[comps[2]+":"+comps[3]] = status
		} else {
			return nil, fmt.Errorf("Invalid key: %v", err)
		}
	}

	return statuses, nil
}
