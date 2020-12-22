// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"

	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/store"
)

const bufferSize = 1024

// NewBlobStore returns a new store service implementation
func NewBlobStore() store.BlobStore {
	return &blob{}
}

type blob struct {
	client pb.BlobStoreService
}

func (b *blob) Read(key string, opts ...store.BlobOption) (io.Reader, error) {
	// validate the key
	if len(key) == 0 {
		return nil, store.ErrMissingKey
	}

	// parse the options
	var options store.BlobOptions
	for _, o := range opts {
		o(&options)
	}

	// execute the rpc
	stream, err := b.cli().Read(context.TODO(), &pb.BlobReadRequest{
		Key: key,
		Options: &pb.BlobOptions{
			Namespace: options.Namespace,
		},
	}, client.WithAuthToken())

	// handle the error
	if verr := errors.FromErr(err); verr != nil && verr.Code == http.StatusNotFound {
		return nil, store.ErrNotFound
	} else if verr != nil {
		return nil, verr
	} else if err != nil {
		return nil, err
	}

	// create a buffer to store the bytes in
	buf := bytes.NewBuffer(nil)

	// keep recieving bytes from the stream until it's closed by the server
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		buf.Write(res.Blob)
	}

	// return the bytes
	return buf, nil
}

func (b *blob) Write(key string, blob io.Reader, opts ...store.BlobOption) error {
	// validate the key
	if len(key) == 0 {
		return store.ErrMissingKey
	}

	// parse the options
	var options store.BlobOptions
	for _, o := range opts {
		o(&options)
	}

	// setup a context
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// open the stream
	stream, err := b.cli().Write(ctx, client.WithAuthToken())
	if verr := errors.FromErr(err); verr != nil {
		return verr
	} else if err != nil {
		return err
	}

	// read from the blob and stream it to the server
	buffer := make([]byte, bufferSize)
	for {
		num, err := blob.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		req := &pb.BlobWriteRequest{
			Key: key,
			Options: &pb.BlobOptions{
				Namespace: options.Namespace,
			},
			Blob: buffer[:num],
		}

		if err := stream.Send(req); err != nil {
			return err
		}
	}

	// wait for the server to process the blob
	_, err = stream.CloseAndRecv()
	return err
}

func (b *blob) Delete(key string, opts ...store.BlobOption) error {
	// validate the key
	if len(key) == 0 {
		return store.ErrMissingKey
	}

	// parse the options
	var options store.BlobOptions
	for _, o := range opts {
		o(&options)
	}

	// execute the rpc
	_, err := b.cli().Delete(context.TODO(), &pb.BlobDeleteRequest{
		Key: key,
		Options: &pb.BlobOptions{
			Namespace: options.Namespace,
		},
	}, client.WithAuthToken())

	// handle the error
	if verr := errors.FromErr(err); verr != nil && verr.Code == http.StatusNotFound {
		return store.ErrNotFound
	} else if verr != nil {
		return verr
	}

	return err
}

func (b *blob) cli() pb.BlobStoreService {
	if b.client == nil {
		b.client = pb.NewBlobStoreService("store", client.DefaultClient)
	}
	return b.client
}
