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
	"strings"
	"time"

	"github.com/google/uuid"

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/namespace"
)

var (
	// eventTTL is the duration events will perist in the store before expiring
	eventTTL = time.Minute * 10
	// eventPollFrequency is the max frequency the manager will check for new events in the store
	eventPollFrequency = time.Minute
)

// eventPrefix is prefixed to the key for event records
const eventPrefix = "event/"

// publishEvent will write the event to the global store and immediately process the event
func (m *manager) publishEvent(eType runtime.EventType, svc *runtime.Service, opts *runtime.CreateOptions) error {
	e := &runtime.Event{
		ID:      uuid.New().String(),
		Type:    eType,
		Service: svc,
		Options: opts,
	}

	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	record := &store.Record{
		Key:    eventPrefix + e.ID,
		Value:  bytes,
		Expiry: eventTTL,
	}

	if err := m.options.Store.Write(record); err != nil {
		return err
	}

	go m.processEvent(record.Key)
	return nil
}

// watchEvents polls the store for events periodically and processes them if they have not already
// done so
func (m *manager) watchEvents() {
	ticker := time.NewTicker(eventPollFrequency)

	for {
		// get the keys of the events
		events, err := m.options.Store.Read(eventPrefix, store.ReadPrefix())
		if err != nil {
			log.Warn("Error listing events: %v", err)
			continue
		}

		// loop through every event
		for _, ev := range events {
			log.Debugf("Process Event: %v", ev.Key)
			m.processEvent(ev.Key)
		}

		<-ticker.C
	}
}

// processEvent will take an event key, verify it hasn't been consumed and then execute it. We pass
// the key and not the ID since the global store and the memory store use the same key prefix so there
// is not point stripping and then re-prefixing.
func (m *manager) processEvent(key string) {
	// check to see if the event has been processed before
	if _, err := m.cache.Read(key); err != store.ErrNotFound {
		return
	}

	// lookup the event
	recs, err := m.options.Store.Read(key)
	if err != nil {
		log.Warnf("Error finding event %v: %v", key, err)
		return
	}
	var ev *runtime.Event
	if err := json.Unmarshal(recs[0].Value, &ev); err != nil {
		log.Warnf("Error unmarshaling event %v: %v", key, err)
	}

	// determine the namespace
	ns := namespace.DefaultNamespace
	if ev.Options != nil && len(ev.Options.Namespace) > 0 {
		ns = ev.Options.Namespace
	}

	// log the event
	log.Infof("Processing %v event for service %v:%v in namespace %v", ev.Type, ev.Service.Name, ev.Service.Version, ns)

	// apply the event to the managed runtime
	switch ev.Type {
	case runtime.Delete:
		err = m.Runtime.Delete(ev.Service, runtime.DeleteNamespace(ns))
	case runtime.Update:
		err = m.Runtime.Update(ev.Service, runtime.UpdateNamespace(ns))
	case runtime.Create:
		err = m.Runtime.Create(ev.Service,
			runtime.CreateImage(ev.Options.Image),
			runtime.CreateType(ev.Options.Type),
			runtime.CreateNamespace(ns),
			runtime.WithArgs(ev.Options.Args...),
			runtime.WithCommand(ev.Options.Command...),
			runtime.WithEnv(m.runtimeEnv(ev.Options)),
		)
	}

	// if there was an error update the status in the cache
	if err != nil {
		log.Warnf("Error processing %v event for service %v:%v in namespace %v: %v", ev.Type, ev.Service.Name, ev.Service.Version, ns, err)
		ev.Service.Metadata = map[string]string{"status": "error", "error": err.Error()}
		m.cacheStatus(ns, ev.Service)
	} else if ev.Type != runtime.Delete {
		m.cacheStatus(ns, ev.Service)
	}

	// write to the store indicating the event has been consumed. We double the ttl to safely know the
	// event will expire before this record
	m.cache.Write(&store.Record{Key: key, Expiry: eventTTL * 2})
}

// runtimeEnv returns the environment variables which should  be used when creating a service.
func (m *manager) runtimeEnv(options *runtime.CreateOptions) []string {
	setEnv := func(p []string, env map[string]string) {
		for _, v := range p {
			parts := strings.Split(v, "=")
			if len(parts) <= 1 {
				continue
			}
			env[parts[0]] = strings.Join(parts[1:], "=")
		}
	}

	// overwrite any values
	env := map[string]string{}

	// set the env vars provided
	setEnv(options.Env, env)

	// override with vars from the Profile
	setEnv(m.options.Profile, env)

	// temp: set the auth namespace. this will be removed once he namespace can be determined from certs.
	if len(options.Namespace) > 0 {
		env["VINE_AUTH_NAMESPACE"] = options.Namespace
	}

	// create a new env
	var vars []string
	for k, v := range env {
		vars = append(vars, k+"="+v)
	}

	// setup the runtime env
	return vars
}
