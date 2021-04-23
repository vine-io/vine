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

package debug

import (
	"sync"
	"time"

	"github.com/lack-io/vine/core/registry"
	"github.com/lack-io/vine/core/registry/cache"
	"github.com/lack-io/vine/lib/cmd"
	log "github.com/lack-io/vine/lib/logger"
	regpb "github.com/lack-io/vine/proto/apis/registry"
)

// Stats is the Debug.Stats handler
type cached struct {
	registry registry.Registry

	sync.RWMutex
	serviceCache []*regpb.Service
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

func (c *cached) services() []*regpb.Service {
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

	serviceMap := make(map[string]*regpb.Service)

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
	var serviceList []*regpb.Service

	for _, service := range serviceMap {
		serviceList = append(serviceList, service)
	}

	// save the list
	c.Lock()
	c.serviceCache = serviceList
	c.Unlock()
	return nil
}
