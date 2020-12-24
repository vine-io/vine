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

// Package service implements the store service interface
package grpc

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/lack-io/vine/proto/errors"
	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/context/metadata"
)

type gRPCStore struct {
	options store.Options

	// The database to use
	Database string

	// The table to use
	Table string

	// Addresses of the nodes
	Nodes []string

	// store service client
	Client pb.StoreService
}

func (s *gRPCStore) Close() error {
	return nil
}

func (s *gRPCStore) Init(opts ...store.Option) error {
	for _, o := range opts {
		o(&s.options)
	}
	s.Database = s.options.Database
	s.Table = s.options.Table
	s.Nodes = s.options.Nodes

	return nil
}

func (s *gRPCStore) Context() context.Context {
	ctx := context.Background()
	md := make(metadata.Metadata)
	if len(s.Database) > 0 {
		md["Vine-Database"] = s.Database
	}

	if len(s.Table) > 0 {
		md["Vine-Table"] = s.Table
	}
	return metadata.NewContext(ctx, md)
}

// Sync all the known records
func (s *gRPCStore) List(opts ...store.ListOption) ([]string, error) {
	options := store.ListOptions{
		Database: s.Database,
		Table:    s.Table,
	}

	for _, o := range opts {
		o(&options)
	}

	listOpts := &pb.ListOptions{
		Database: options.Database,
		Table:    options.Table,
		Prefix:   options.Prefix,
		Suffix:   options.Suffix,
		Limit:    uint64(options.Limit),
		Offset:   uint64(options.Offset),
	}

	stream, err := s.Client.List(s.Context(), &pb.ListRequest{Options: listOpts}, client.WithAddress(s.Nodes...))
	if err != nil && errors.Equal(err, errors.NotFound("", "")) {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	defer stream.Close()

	var keys []string

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return keys, err
		}

		for _, key := range rsp.Keys {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Read a record with key
func (s *gRPCStore) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	options := store.ReadOptions{
		Database: s.Database,
		Table:    s.Table,
	}

	for _, o := range opts {
		o(&options)
	}

	readOpts := &pb.ReadOptions{
		Database: options.Database,
		Table:    options.Table,
		Prefix:   options.Prefix,
		Suffix:   options.Suffix,
		Limit:    uint64(options.Limit),
		Offset:   uint64(options.Offset),
	}

	rsp, err := s.Client.Read(s.Context(), &pb.ReadRequest{
		Key:     key,
		Options: readOpts,
	}, client.WithAddress(s.Nodes...))
	if err != nil && errors.Equal(err, errors.NotFound("", "")) {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	records := make([]*store.Record, 0, len(rsp.Records))

	for _, val := range rsp.Records {
		metadata := make(map[string]interface{})

		for k, v := range val.Metadata {
			switch v.Type {
			// TODO: parse all types
			default:
				metadata[k] = v
			}
		}

		records = append(records, &store.Record{
			Key:      val.Key,
			Value:    val.Value,
			Expiry:   time.Duration(val.Expiry) * time.Second,
			Metadata: metadata,
		})
	}

	return records, nil
}

// Write a record
func (s *gRPCStore) Write(record *store.Record, opts ...store.WriteOption) error {
	options := store.WriteOptions{
		Database: s.Database,
		Table:    s.Table,
	}

	for _, o := range opts {
		o(&options)
	}

	writeOpts := &pb.WriteOptions{
		Database: options.Database,
		Table:    options.Table,
	}

	metadata := make(map[string]*pb.Field)

	for k, v := range record.Metadata {
		metadata[k] = &pb.Field{
			Type:  reflect.TypeOf(v).String(),
			Value: fmt.Sprintf("%v", v),
		}
	}

	_, err := s.Client.Write(s.Context(), &pb.WriteRequest{
		Record: &pb.Record{
			Key:      record.Key,
			Value:    record.Value,
			Expiry:   int64(record.Expiry.Seconds()),
			Metadata: metadata,
		},
		Options: writeOpts}, client.WithAddress(s.Nodes...))
	if err != nil && errors.Equal(err, errors.NotFound("", "")) {
		return store.ErrNotFound
	}

	return err
}

// Delete a record with key
func (s *gRPCStore) Delete(key string, opts ...store.DeleteOption) error {
	options := store.DeleteOptions{
		Database: s.Database,
		Table:    s.Table,
	}

	for _, o := range opts {
		o(&options)
	}

	deleteOpts := &pb.DeleteOptions{
		Database: options.Database,
		Table:    options.Table,
	}

	_, err := s.Client.Delete(s.Context(), &pb.DeleteRequest{
		Key:     key,
		Options: deleteOpts,
	}, client.WithAddress(s.Nodes...))
	if err != nil && errors.Equal(err, errors.NotFound("", "")) {
		return store.ErrNotFound
	}

	return err
}

func (s *gRPCStore) String() string {
	return "grpc"
}

func (s *gRPCStore) Options() store.Options {
	return s.options
}

// NewStore returns a new store service implementation
func NewStore(opts ...store.Option) store.Store {
	var options store.Options
	for _, o := range opts {
		o(&options)
	}

	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	service := &gRPCStore{
		options:  options,
		Database: options.Database,
		Table:    options.Table,
		Nodes:    options.Nodes,
		Client:   pb.NewStoreService("go.vine.store", options.Client),
	}

	return service
}
