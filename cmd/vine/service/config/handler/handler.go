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

package handler

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/config"
	"github.com/lack-io/vine/service/client"
	cr "github.com/lack-io/vine/service/config/reader"
	jr "github.com/lack-io/vine/service/config/reader/json"
	"github.com/lack-io/vine/service/config/source"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/namespace"
)

var (
	PathSplitter = "."
	WatchTopic   = "go.vine.config.events"
	watchers     = make(map[string][]*watcher)

	// we now support json only
	reader = jr.NewReader()
	mtx    sync.RWMutex
)

type Config struct {
	Store store.Store
}

// setNamespace figures out what the namespace should be
func setNamespace(ctx context.Context, v string) string {
	ns := namespace.FromContext(ctx)
	if ns == "go.vine" {
		ns = "vine"
	}

	return ns + ":" + v
}

func (c *Config) Read(ctx context.Context, req *pb.ReadRequest, rsp *pb.ReadResponse) error {
	if len(req.Namespace) == 0 {
		return errors.BadRequest("go.vine.config.Read", "invalid id")
	}

	namespace := setNamespace(ctx, req.Namespace)

	ch, err := c.Store.Read(namespace)
	if err == store.ErrNotFound {
		return errors.NotFound("go.vine.config.Read", "Not found")
	} else if err != nil {
		return errors.BadRequest("go.vine.config.Read", "read error: %v: %v", err, req.Namespace)
	}

	rsp.Change = new(pb.Change)

	// Unmarshal value
	if err = json.Unmarshal(ch[0].Value, rsp.Change); err != nil {
		return errors.BadRequest("go.vine.config.Read", "unmarshal value error: %v", err)
	}

	// if dont need path, we return all of the data
	if len(req.Path) == 0 {
		return nil
	}

	rc := rsp.Change.ChangeSet

	// generate reader.Values from the changeset
	values, err := values(&source.ChangeSet{
		Timestamp: time.Unix(rc.Timestamp, 0),
		Data:      []byte(rc.Data),
		CheckSum:  rc.Checksum,
		Format:    rc.Format,
		Source:    rc.Source,
	})
	if err != nil {
		return errors.InternalServerError("go.vine.config.Read", err.Error())
	}

	// peel apart the path
	parts := strings.Split(req.Path, PathSplitter)

	// we just want to pass back bytes
	rsp.Change.ChangeSet.Data = string(values.Get(parts...).Bytes())

	return nil
}

func (c *Config) Create(ctx context.Context, req *pb.CreateRequest, rsp *pb.CreateResponse) error {
	if req.Change == nil || req.Change.ChangeSet == nil {
		return errors.BadRequest("go.vine.config.Create", "invalid change")
	}

	if len(req.Change.Namespace) == 0 {
		return errors.BadRequest("go.vine.config.Create", "invalid id")
	}

	if len(req.Change.Path) > 0 {
		vals, err := values(&source.ChangeSet{
			Format: "json",
		})
		if err != nil {
			return errors.InternalServerError("go.vine.config.Create", err.Error())
		}

		// peel apart the path
		parts := strings.Split(req.Change.Path, PathSplitter)
		// set the values
		vals.Set(req.Change.ChangeSet.Data, parts...)
		// change the changeset value
		req.Change.ChangeSet.Data = string(vals.Bytes())
	}

	req.Change.ChangeSet.Timestamp = time.Now().Unix()

	namespace := setNamespace(ctx, req.Change.Namespace)

	record := &store.Record{
		Key: namespace,
	}

	var err error
	record.Value, err = json.Marshal(req.Change)
	if err != nil {
		return errors.BadRequest("go.vine.config.Create", "marshal error: %v", err)
	}

	if err := c.Store.Write(record); err != nil {
		return errors.BadRequest("go.vine.config.Create", "create new into db error: %v", err)
	}

	_ = publish(ctx, &pb.WatchResponse{Namespace: namespace, ChangeSet: req.Change.ChangeSet})

	return nil
}

func (c *Config) Update(ctx context.Context, req *pb.UpdateRequest, rsp *pb.UpdateResponse) error {
	if req.Change == nil || req.Change.ChangeSet == nil {
		return errors.BadRequest("go.vine.config.Update", "invalid change")
	}

	if len(req.Change.Namespace) == 0 {
		return errors.BadRequest("go.vine.config.Update", "invalid id")
	}

	// set the changeset timestamp
	req.Change.ChangeSet.Timestamp = time.Now().Unix()

	oldCh := &pb.Change{}

	namespace := setNamespace(ctx, req.Change.Namespace)

	// Get the current change set
	var record *store.Record
	records, err := c.Store.Read(namespace)
	if err != nil {
		if err.Error() != "not found" {
			return errors.BadRequest("go.vine.config.Update", "read old value error: %v", err)
		}
		// create new record
		record = new(store.Record)
		record.Key = namespace
	} else {
		// Unmarshal value
		if err := json.Unmarshal(records[0].Value, oldCh); err != nil {
			return errors.BadRequest("go.vine.config.Read", "unmarshal value error: %v", err)
		}
		record = records[0]
	}

	// generate a new base changeset
	changeSet := &source.ChangeSet{
		Format: "json",
		Data:   []byte(`{}`),
	}

	if oldCh.ChangeSet != nil {
		changeSet = &source.ChangeSet{
			Timestamp: time.Unix(oldCh.ChangeSet.Timestamp, 0),
			Data:      []byte(oldCh.ChangeSet.Data),
			CheckSum:  oldCh.ChangeSet.Checksum,
			Source:    oldCh.ChangeSet.Source,
			Format:    oldCh.ChangeSet.Format,
		}
	}

	var newChange *source.ChangeSet

	// Set the change at a particular path
	if len(req.Change.Path) > 0 {
		// Get values from existing change
		values, err := values(changeSet)
		if err != nil {
			return errors.InternalServerError("go.vine.config.Update", "error getting existing change: %v", err)
		}

		// Apply the data to the existing change
		values.Set(req.Change.ChangeSet.Data, strings.Split(req.Change.Path, PathSplitter)...)

		// Create a new change
		newChange, err = merge(&source.ChangeSet{Data: values.Bytes()})
		if err != nil {
			return errors.InternalServerError("go.vine.config.Update", "create a new change error: %v", err)
		}
	} else {
		// No path specified, business as usual
		newChange, err = merge(changeSet, &source.ChangeSet{
			Timestamp: time.Unix(req.Change.ChangeSet.Timestamp, 0),
			Data:      []byte(req.Change.ChangeSet.Data),
			CheckSum:  req.Change.ChangeSet.Checksum,
			Source:    req.Change.ChangeSet.Source,
			Format:    req.Change.ChangeSet.Format,
		})
		if err != nil {
			return errors.BadRequest("go.vine.svc.config.Update", "merge all error: %v", err)
		}
	}

	// update change set
	req.Change.ChangeSet = &pb.ChangeSet{
		Timestamp: newChange.Timestamp.Unix(),
		Data:      string(newChange.Data),
		Checksum:  newChange.CheckSum,
		Source:    newChange.Source,
		Format:    newChange.Format,
	}

	record.Value, err = json.Marshal(req.Change)
	if err != nil {
		return errors.BadRequest("go.vine.config.Update", "marshal error: %v", err)
	}

	if err := c.Store.Write(record); err != nil {
		return errors.BadRequest("go.vine.config.Update", "update into db error: %v", err)
	}

	_ = publish(ctx, &pb.WatchResponse{Namespace: namespace, ChangeSet: req.Change.ChangeSet})

	return nil
}

func (c *Config) Delete(ctx context.Context, req *pb.DeleteRequest, rsp *pb.DeleteResponse) error {
	if req.Change == nil {
		return errors.BadRequest("go.vine.svc.Delete", "invalid change")
	}

	if len(req.Change.Namespace) == 0 {
		return errors.BadRequest("go.vine.svc.Delete", "invalid id")
	}

	if req.Change.ChangeSet == nil {
		req.Change.ChangeSet = &pb.ChangeSet{}
	}

	req.Change.ChangeSet.Timestamp = time.Now().Unix()

	namespace := setNamespace(ctx, req.Change.Namespace)

	// We're going to delete the record as we have no path and no data
	if len(req.Change.Path) == 0 {
		if err := c.Store.Delete(namespace); err != nil {
			return errors.BadRequest("go.vine.svc.Delete", "delete from db error: %v", err)
		}
		return nil
	}

	// We've got a path. Let's update the required path

	// Get the current change set
	records, err := c.Store.Read(namespace)
	if err != nil {
		if err.Error() != "not found" {
			return errors.BadRequest("go.vine.svc.Delete", "read old value error: %v", err)
		}
		return nil
	}

	ch := &pb.Change{}
	// Unmarshal value
	if err := json.Unmarshal(records[0].Value, ch); err != nil {
		return errors.BadRequest("go.vine.config.Read", "unmarshal value error: %v", err)
	}

	// Get the current config as values
	values, err := values(&source.ChangeSet{
		Timestamp: time.Unix(ch.ChangeSet.Timestamp, 0),
		Data:      []byte(ch.ChangeSet.Data),
		CheckSum:  ch.ChangeSet.Checksum,
		Source:    ch.ChangeSet.Source,
		Format:    ch.ChangeSet.Format,
	})
	if err != nil {
		return errors.BadRequest("go.vine.svc.Delete", "Get the current config as values error: %v", err)
	}

	// Delete at the given path
	values.Del(strings.Split(req.Change.Path, PathSplitter)...)

	// Create a change record from the values
	change, err := merge(&source.ChangeSet{Data: values.Bytes()})
	if err != nil {
		return errors.BadRequest("go.vine.svc.Delete", "Create a change record from the values error: %v", err)
	}

	// Update change set
	req.Change.ChangeSet = &pb.ChangeSet{
		Timestamp: change.Timestamp.Unix(),
		Data:      string(change.Data),
		Checksum:  change.CheckSum,
		Format:    change.Format,
		Source:    change.Source,
	}

	records[0].Value, err = json.Marshal(req.Change)
	if err != nil {
		return errors.BadRequest("go.vine.config.Update", "marshal error: %v", err)
	}

	if err := c.Store.Write(records[0]); err != nil {
		return errors.BadRequest("go.vine.svc.Delete", "update record set to db error: %v", err)
	}

	_ = publish(ctx, &pb.WatchResponse{Namespace: namespace, ChangeSet: req.Change.ChangeSet})

	return nil
}

func (c *Config) List(ctx context.Context, req *pb.ListRequest, rsp *pb.ListResponse) (err error) {
	list, err := c.Store.List()
	if err != nil {
		return errors.BadRequest("go.vine.config.List", "query value error: %v", err)
	}

	ns := setNamespace(ctx, "")

	// TODO: optimise filtering for prefix listing
	for _, v := range list {
		if !strings.HasPrefix(v, ns) {
			continue
		}

		rec, err := c.Store.Read(v)
		if err != nil {
			return err
		}

		ch := &pb.Change{}
		if err := json.Unmarshal(rec[0].Value, ch); err != nil {
			return errors.BadRequest("go.vine.config.Read", "unmarshal value error: %v", err)
		}

		if ch.ChangeSet != nil {
			ch.ChangeSet.Data = string(ch.ChangeSet.Data)
		}

		rsp.Values = append(rsp.Values, ch)
	}

	return nil
}

func (c *Config) Watch(ctx context.Context, req *pb.WatchRequest, stream pb.Config_WatchStream) error {
	if len(req.Namespace) == 0 {
		return errors.BadRequest("go.vine.svc.Watch", "invalid id")
	}

	namespace := setNamespace(ctx, req.Namespace)

	watch, err := Watch(namespace)
	if err != nil {
		return errors.BadRequest("go.vine.svc.Watch", "watch error: %v", err)
	}
	defer watch.Stop()

	go func() {
		select {
		case <-ctx.Done():
			watch.Stop()
			stream.Close()
		}
	}()

	for {
		ch, err := watch.Next()
		if err != nil {
			return errors.BadRequest("go.vine.svc.Watch", "listen the Next error: %v", err)
		}
		if ch.ChangeSet != nil {
			ch.ChangeSet.Data = string(ch.ChangeSet.Data)
		}
		if err := stream.Send(ch); err != nil {
			return errors.BadRequest("go.vine.svc.Watch", "send the Change error: %v", err)
		}
	}
}

// Used as a subscriber between config services for events
func Watcher(ctx context.Context, ch *pb.WatchResponse) error {
	mtx.RLock()
	for _, sub := range watchers[ch.Namespace] {
		select {
		case sub.next <- ch:
		case <-time.After(time.Millisecond * 100):
		}
	}
	mtx.RUnlock()
	return nil
}

func merge(ch ...*source.ChangeSet) (*source.ChangeSet, error) {
	return reader.Merge(ch...)
}

func values(ch *source.ChangeSet) (cr.Values, error) {
	return reader.Values(ch)
}

// publish a change
func publish(ctx context.Context, ch *pb.WatchResponse) error {
	req := client.NewMessage(WatchTopic, ch)
	return client.Publish(ctx, req)
}
