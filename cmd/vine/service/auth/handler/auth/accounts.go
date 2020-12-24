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

package auth

import (
	"context"
	"encoding/json"
	"strings"

	pb "github.com/lack-io/vine/proto/auth"
	"github.com/lack-io/vine/proto/errors"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/store"
	"github.com/lack-io/vine/util/namespace"
)

// List returns all auth accounts
func (a *Auth) List(ctx context.Context, req *pb.ListAccountsRequest, rsp *pb.ListAccountsResponse) error {
	// setup the defaults incase none exist
	a.setupDefaultAccount(namespace.FromContext(ctx))

	// get the records from the store
	key := strings.Join([]string{storePrefixAccounts, namespace.FromContext(ctx), ""}, joinKey)
	recs, err := a.Options.Store.Read(key, store.ReadPrefix())
	if err != nil {
		return errors.InternalServerError("go.vine.auth", "Unable to read from store: %v", err)
	}

	// unmarshal the records
	var accounts = make([]*auth.Account, 0, len(recs))
	for _, rec := range recs {
		var r *auth.Account
		if err := json.Unmarshal(rec.Value, &r); err != nil {
			return errors.InternalServerError("go.vine.auth", "Error to unmarshaling json: %v. Value: %v", err, string(rec.Value))
		}
		accounts = append(accounts, r)
	}

	// serialize the accounts
	rsp.Accounts = make([]*pb.Account, 0, len(recs))
	for _, a := range accounts {
		rsp.Accounts = append(rsp.Accounts, serializeAccount(a))
	}

	return nil
}

func serializeAccount(a *auth.Account) *pb.Account {
	return &pb.Account{
		Id:       a.ID,
		Type:     a.Type,
		Scopes:   a.Scopes,
		Issuer:   a.Issuer,
		Metadata: a.Metadata,
	}
}
