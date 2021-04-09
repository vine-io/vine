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

package rules

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/auth"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/store"
	memStore "github.com/lack-io/vine/service/store/memory"
	"github.com/lack-io/vine/util/namespace"
)

const (
	storePrefixRules = "rules"
	joinKey          = "/"
)

var defaultRule = &auth.Rule{
	ID:    "default",
	Scope: auth.ScopePublic,
	Resource: &auth.Resource{
		Type:     "*",
		Name:     "*",
		Endpoint: "*",
	},
}

// Rules processes RPC calls
type Rules struct {
	Options auth.Options

	namespaces map[string]bool
	sync.Mutex
}

// Init the auth
func (r *Rules) Init(opts ...auth.Option) {
	for _, o := range opts {
		o(&r.Options)
	}

	// use the default store as a fallback
	if r.Options.Store == nil {
		r.Options.Store = store.DefaultStore
	}

	// noop will not work for auth
	if r.Options.Store.String() == "noop" {
		r.Options.Store = memStore.NewStore()
	}
}

func (r *Rules) setupDefaultRules(ns string) {
	r.Lock()
	defer r.Unlock()

	// setup the namespace cache if not yet done
	if r.namespaces == nil {
		r.namespaces = make(map[string]bool)
	}

	// check to see if the default rule has already been verified
	if _, ok := r.namespaces[ns]; ok {
		return
	}

	// setup a context with the namespace
	ctx := namespace.ContextWithNamespace(context.TODO(), ns)

	// check to see if we need to create the default account
	key := strings.Join([]string{storePrefixRules, ns, ""}, joinKey)
	recs, err := r.Options.Store.Read(key, store.ReadPrefix())
	if err != nil {
		return
	}

	// create the account if none exist in the namespace
	if len(recs) == 0 {
		req := &pb.CreateRequest{
			Rule: &pb.Rule{
				Id:     defaultRule.ID,
				Scope:  defaultRule.Scope,
				Access: pb.Access_GRANTED,
				Resource: &pb.Resource{
					Type:     defaultRule.Resource.Type,
					Name:     defaultRule.Resource.Name,
					Endpoint: defaultRule.Resource.Endpoint,
				},
			},
		}

		r.Create(ctx, req, &pb.CreateResponse{})
	}

	// set the namespace in the cache
	r.namespaces[ns] = true
}

// Create a rule giving a scope access to a resource
func (r *Rules) Create(ctx context.Context, req *pb.CreateRequest, rsp *pb.CreateResponse) error {
	// Validate the request
	if req.Rule == nil {
		return errors.BadRequest("go.vine.auth", "Rule missing")
	}
	if len(req.Rule.Id) == 0 {
		return errors.BadRequest("go.vine.auth", "ID missing")
	}
	if req.Rule.Resource == nil {
		return errors.BadRequest("go.vine.auth", "Resource missing")
	}
	if req.Rule.Access == pb.Access_UNKNOWN {
		return errors.BadRequest("go.vine.auth", "Access missing")
	}

	// Chck the rule doesn't exist
	ns := namespace.FromContext(ctx)
	key := strings.Join([]string{storePrefixRules, ns, req.Rule.Id}, joinKey)
	if _, err := r.Options.Store.Read(key); err == nil {
		return errors.BadRequest("go.vine.auth", "A rule with this ID already exists")
	}

	// Encode the rule
	bytes, err := json.Marshal(req.Rule)
	if err != nil {
		return errors.InternalServerError("go.vine.auth", "Unable to marshal rule: %v", err)
	}

	// Write to the store
	if err := r.Options.Store.Write(&store.Record{Key: key, Value: bytes}); err != nil {
		return errors.InternalServerError("go.vine.auth", "Unable to write to the store: %v", err)
	}

	return nil
}

// Delete a scope access to a resource
func (r *Rules) Delete(ctx context.Context, req *pb.DeleteRequest, rsp *pb.DeleteResponse) error {
	// Validate the request
	if len(req.Id) == 0 {
		return errors.BadRequest("go.vine.auth", "ID missing")
	}

	// Delete the rule
	ns := namespace.FromContext(ctx)
	key := strings.Join([]string{storePrefixRules, ns, req.Id}, joinKey)
	err := r.Options.Store.Delete(key)
	if err == store.ErrNotFound {
		return errors.BadRequest("go.vine.auth", "Rule not found")
	} else if err != nil {
		return errors.InternalServerError("go.vine.auth", "Unable to delete key from store: %v", err)
	}

	return nil
}

// List returns all the rules
func (r *Rules) List(ctx context.Context, req *pb.ListRequest, rsp *pb.ListResponse) error {
	// setup the defaults incase none exist
	r.setupDefaultRules(namespace.FromContext(ctx))

	// get the records from the store
	ns := namespace.FromContext(ctx)
	prefix := strings.Join([]string{storePrefixRules, ns, ""}, joinKey)
	recs, err := r.Options.Store.Read(prefix, store.ReadPrefix())
	if err != nil {
		return errors.InternalServerError("go.vine.auth", "Unable to read from store: %v", err)
	}

	// unmarshal the records
	rsp.Rules = make([]*pb.Rule, 0, len(recs))
	for _, rec := range recs {
		var r *pb.Rule
		if err := json.Unmarshal(rec.Value, &r); err != nil {
			return errors.InternalServerError("go.vine.auth", "Error to unmarshaling json: %v. Value: %v", err, string(rec.Value))
		}
		rsp.Rules = append(rsp.Rules, r)
	}

	return nil
}
