// Copyright 2020 lack
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

package debug

import (
	"sync"
	"time"

	"github.com/lack-io/vine/service/config/cmd"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/cache"
)

// Stats is the Debug.Stats handler
type cached struct {
	registry registry.Registry

	sync.RWMutex
	serviceCache []*registry.Service
}

func newCache(done <-chan bool) *cached {
	c := &cached{
		registry: cache.New(*cmd.DefaultOptions().Registry),
	}

	// first scan
	if err := c.scan(); err != nil {
		return nil
	}

	go c.run(done)

	return c
}

func (c *cached) services() []*registry.Service {
	c.RLock()
	defer c.RUnlock()
	return c.serviceCache
}

func (c *cached) run(done <-chan bool) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-done:
			return
		case <-t.C:
			if err := c.scan(); err != nil {
				log.Debug(err)
			}
		}
	}
}

func (c *cached) scan() error {
	services, err := c.registry.ListServices()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	serviceMap := make(map[string]*registry.Service)

	// check each service has nodes
	for _, service := range services {
		if len(service.Nodes) > 0 {
			serviceMap[service.Name+service.Version] = service
			continue
		}

		// get nodes that does not exist
		newServices, err := c.registry.GetService(service.Name)
		if err != nil {
			continue
		}

		// store service by version
		for _, service := range newServices {
			serviceMap[service.Name+service.Version] = service
		}
	}

	// flatten the map
	var serviceList []*registry.Service

	for _, service := range serviceMap {
		serviceList = append(serviceList, service)
	}

	// save the list
	c.Lock()
	c.serviceCache = serviceList
	c.Unlock()
	return nil
}
