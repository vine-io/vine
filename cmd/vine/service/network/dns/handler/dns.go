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
	"errors"

	"github.com/lack-io/vine/cmd/vine/service/network/dns/provider"
	log "github.com/lack-io/vine/lib/logger"
	dns "github.com/lack-io/vine/proto/services/network/dns"
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
