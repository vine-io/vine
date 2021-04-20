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

package handler

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/lack-io/vine/lib/store"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/store"
	"github.com/lack-io/vine/util/context/metadata"
	"github.com/lack-io/vine/util/namespace"
)

type Store struct {
	// The default store
	Default store.Store

	// Store initialiser
	New func(string, string) (store.Store, error)

	// Store map
	sync.RWMutex

	Stores map[string]bool
}

// TODO: remove this horrible bs
func (s *Store) get(ctx context.Context, database, table string) (string, string) {
	// lock (might be a race)
	s.Lock()
	defer s.Unlock()

	// get the namespace from context
	ns := namespace.FromContext(ctx)
	// we're using "vine" as the database"
	// TODO: change default namespace to vine
	if ns == "go.vine" {
		ns = "vine"
	}

	// retrieve values from metadata
	// TODO: switch to options
	if md, ok := metadata.FromContext(ctx); ok {
		// TODO: remove this, its here only for legacy purposes
		if v, ok := md.Get("Vine-Database"); ok && len(v) > 0 {
			database = v
		}
		if v, ok := md.Get("Vine-Table"); ok && len(v) > 0 {
			table = v
		}
	}

	// set the database to the namespace
	if len(ns) > 0 {
		database = ns
	}

	// reset database to options if not set
	if len(database) == 0 {
		database = s.Default.Options().Database
	}

	// reset table to options if not set
	if len(table) == 0 {
		table = s.Default.Options().Table
	}

	// just use the default if nothing is specified
	if len(database) == 0 && len(table) == 0 {
		return "vine", "store"
	}

	// attempt to get the database
	_, ok := s.Stores[database+":"+table]
	if !ok {
		// set that we know about this database/table
		s.New(database, table)
	}

	// save store
	s.Stores[database+":"+table] = true

	return database, table
}

func (s *Store) Read(ctx context.Context, req *pb.ReadRequest, rsp *pb.ReadResponse) error {
	var opts []store.ReadOption
	var database, table string

	if req.Options != nil {
		if req.Options.Prefix {
			opts = append(opts, store.ReadPrefix())
		}
		if db := req.Options.Database; len(db) > 0 {
			database = db
		}
		if tb := req.Options.Table; len(tb) > 0 {
			table = tb
		}
	}

	// get new store
	database, table = s.get(ctx, database, table)
	opts = append(opts, store.ReadFrom(database, table))

	vals, err := s.Default.Read(req.Key, opts...)
	if err != nil && err == store.ErrNotFound {
		return errors.NotFound("go.vine.store", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}

	for _, val := range vals {
		rsp.Records = append(rsp.Records, &pb.Record{
			Key:    val.Key,
			Value:  val.Value,
			Expiry: int64(val.Expiry.Seconds()),
		})
	}
	return nil
}

func (s *Store) Write(ctx context.Context, req *pb.WriteRequest, rsp *pb.WriteResponse) error {
	var database, table string

	if req.Options != nil {
		if db := req.Options.Database; len(db) > 0 {
			database = db
		}
		if tb := req.Options.Table; len(tb) > 0 {
			table = tb
		}
	}

	// get new store
	database, table = s.get(ctx, database, table)

	if req.Record == nil {
		return errors.BadRequest("go.vine.store", "no record specified")
	}

	record := &store.Record{
		Key:    req.Record.Key,
		Value:  req.Record.Value,
		Expiry: time.Duration(req.Record.Expiry) * time.Second,
	}

	var opts []store.WriteOption
	opts = append(opts, store.WriteTo(database, table))

	err := s.Default.Write(record, opts...)
	if err != nil && err == store.ErrNotFound {
		return errors.NotFound("go.vine.store", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, req *pb.DeleteRequest, rsp *pb.DeleteResponse) error {
	var database, table string

	if req.Options != nil {
		if db := req.Options.Database; len(db) > 0 {
			database = db
		}
		if tb := req.Options.Table; len(tb) > 0 {
			table = tb
		}
	}

	// get new store
	database, table = s.get(ctx, database, table)

	var opts []store.DeleteOption
	opts = append(opts, store.DeleteFrom(database, table))

	if err := s.Default.Delete(req.Key, opts...); err == store.ErrNotFound {
		return errors.NotFound("go.vine.store", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}
	return nil
}

// TODO: lock down to admin?
func (s *Store) Databases(ctx context.Context, req *pb.DatabasesRequest, rsp *pb.DatabasesResponse) error {
	recs, err := s.Default.Read("databases/", store.ReadPrefix(), store.ReadFrom("vine", "internal"))
	if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}
	rsp.Databases = make([]string, len(recs))
	for i, r := range recs {
		rsp.Databases[i] = strings.TrimPrefix(r.Key, "databases/")
	}
	return nil
}

// TODO: lock down to admin?
func (s *Store) Tables(ctx context.Context, req *pb.TablesRequest, rsp *pb.TablesResponse) error {
	recs, err := s.Default.Read("tables/"+req.Database+"/", store.ReadPrefix(), store.ReadFrom("vine", "internal"))
	if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}
	rsp.Tables = make([]string, len(recs))
	for i, r := range recs {
		rsp.Tables[i] = strings.TrimPrefix(r.Key, "tables/"+req.Database+"/")
	}
	return nil
}

func (s *Store) List(ctx context.Context, req *pb.ListRequest, stream pb.Store_ListStream) error {
	var database, table string

	if req.Options != nil {
		if db := req.Options.Database; len(db) > 0 {
			database = db
		}
		if tb := req.Options.Table; len(tb) > 0 {
			table = tb
		}
	}

	// get new store
	database, table = s.get(ctx, database, table)

	var opts []store.ListOption
	opts = append(opts, store.ListFrom(database, table))

	vals, err := s.Default.List(opts...)
	if err != nil && err == store.ErrNotFound {
		return errors.NotFound("go.vine.store", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}

	rsp := new(pb.ListResponse)

	// TODO: batch sync
	for _, val := range vals {
		rsp.Keys = append(rsp.Keys, val)
	}

	err = stream.Send(rsp)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return errors.InternalServerError("go.vine.store", err.Error())
	}
	return nil
}
