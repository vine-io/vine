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

	"github.com/pkg/errors"

	"github.com/lack-io/vine/cmd/vine/service/network/dns/provider"
	dns "github.com/lack-io/vine/proto/services/network/dns"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/util/context/metadata"
)

// DNS handles incoming gRPC requests
type DNS struct {
	provider    provider.Provider
	bearerToken string
}

// Advertise adds records of network nodes to DNS
func (d *DNS) Advertise(ctx context.Context, req *dns.AdvertiseRequest, rsp *dns.AdvertiseResponse) error {
	log.Debug("Received Advertise Request")
	if err := d.validateMetadata(ctx); err != nil {
		return err
	}
	return d.provider.Advertise(req.Records...)
}

// Remove removes itself from DNS
func (d *DNS) Remove(ctx context.Context, req *dns.RemoveRequest, rsp *dns.RemoveResponse) error {
	log.Debug("Received Remove Request")
	if err := d.validateMetadata(ctx); err != nil {
		return err
	}
	return d.provider.Remove(req.Records...)
}

// Resolve looks up matching records and returns any matches
func (d *DNS) Resolve(ctx context.Context, req *dns.ResolveRequest, rsp *dns.ResolveResponse) error {
	log.Debugf("Received Resolve Request")
	if err := d.validateMetadata(ctx); err != nil {
		return err
	}
	providerResponse, err := d.provider.Resolve(req.Name, req.Type)
	if err != nil {
		return err
	}
	rsp.Records = providerResponse
	return nil
}

func (d *DNS) validateMetadata(ctx context.Context) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return errors.New("Denied: error getting request metadata")
	}
	token, found := md["Authorization"]
	if !found {
		return errors.New("Denied: Authorization metadata not provided")
	}
	if token != "Bearer "+d.bearerToken {
		return errors.New("Denied: Authorization metadata is not valid")
	}
	return nil
}
