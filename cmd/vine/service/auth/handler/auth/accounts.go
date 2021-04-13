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

package auth

import (
	"context"

	pb "github.com/lack-io/vine/proto/services/auth"
	"github.com/lack-io/vine/service/auth"
)

// List returns all auth accounts
func (a *Auth) List(ctx context.Context, req *pb.ListAccountsRequest, rsp *pb.ListAccountsResponse) error {
	// setup the defaults incase none exist
	//a.setupDefaultAccount(namespace.FromContext(ctx))
	//
	//// get the records from the store
	//key := strings.Join([]string{storePrefixAccounts, namespace.FromContext(ctx), ""}, joinKey)
	//recs, err := a.Options.Store.Read(key, store.ReadPrefix())
	//if err != nil {
	//	return errors.InternalServerError("go.vine.auth", "Unable to read from store: %v", err)
	//}
	//
	//// unmarshal the records
	//var accounts = make([]*auth.Account, 0, len(recs))
	//for _, rec := range recs {
	//	var r *auth.Account
	//	if err := json.Unmarshal(rec.Value, &r); err != nil {
	//		return errors.InternalServerError("go.vine.auth", "Error to unmarshaling json: %v. Value: %v", err, string(rec.Value))
	//	}
	//	accounts = append(accounts, r)
	//}
	//
	//// serialize the accounts
	//rsp.Accounts = make([]*pb.Account, 0, len(recs))
	//for _, a := range accounts {
	//	rsp.Accounts = append(rsp.Accounts, serializeAccount(a))
	//}

	return nil
}

func serializeAccount(a *auth.Account) *pb.Account {
	return &pb.Account{
		Id:       a.ID,
		Type:     string(a.Type),
		Scopes:   a.Scopes,
		Issuer:   a.Issuer,
		Metadata: a.Metadata,
	}
}
