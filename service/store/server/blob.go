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

package server

import (
	"bytes"
	"context"
	"io"

	authns "github.com/lack-io/vine/internal/auth/namespace"
	"github.com/lack-io/vine/internal/namespace"
	pb "github.com/lack-io/vine/proto/store"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/store"
	gostore "github.com/lack-io/vine/service/store"
)

const bufferSize = 1024

type blobHandler struct{}

func (b *blobHandler) Read(ctx context.Context, req *pb.BlobReadRequest, stream pb.BlobStore_ReadStream) error {
	// parse the options
	if ns := req.GetOptions().GetNamespace(); len(ns) == 0 {
		req.Options = &pb.BlobOptions{
			Namespace: namespace.FromContext(ctx),
		}
	}

	// authorize the request
	if err := authns.Authorize(ctx, req.Options.Namespace); err == authns.ErrForbidden {
		return errors.Forbidden("store.Blob.Read", err.Error())
	} else if err == authns.ErrUnauthorized {
		return errors.Unauthorized("store.Blob.Read", err.Error())
	} else if err != nil {
		return errors.InternalServerError("store.Blob.Read", err.Error())
	}

	// execute the request
	blob, err := store.DefaultBlobStore.Read(req.Key, gostore.BlobNamespace(req.Options.Namespace))
	if err == gostore.ErrNotFound {
		return errors.NotFound("store.Blob.Read", "Blob not found")
	} else if err == gostore.ErrMissingKey {
		return errors.BadRequest("store.Blob.Read", "Missing key")
	} else if err != nil {
		return errors.InternalServerError("store.Blob.Read", err.Error())
	}

	// read from the blob and stream it to the client
	buffer := make([]byte, bufferSize)
	for {
		num, err := blob.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := stream.Send(&pb.BlobReadResponse{Blob: buffer[:num]}); err != nil {
			return err
		}
	}

	return stream.Close()
}

func (b *blobHandler) Write(ctx context.Context, stream pb.BlobStore_WriteStream) error {
	// the key and options are passed on each message but we only need to extract them once
	var buf *bytes.Buffer
	var key string
	var options *pb.BlobOptions

	// recieve the blob from the client
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.InternalServerError("store.Blob.Write", err.Error())
		}

		if buf == nil {
			// first message recieved from the stream
			buf = bytes.NewBuffer(req.Blob)
			key = req.Key
			options = req.Options

			// parse the options
			if options == nil || len(options.Namespace) == 0 {
				options = &pb.BlobOptions{Namespace: namespace.FromContext(ctx)}
			}

			// authorize the request. do this inside the loop so we fail fast
			if err := authns.Authorize(ctx, options.Namespace); err == authns.ErrForbidden {
				return errors.Forbidden("store.Blob.Write", err.Error())
			} else if err == authns.ErrUnauthorized {
				return errors.Unauthorized("store.Blob.Write", err.Error())
			} else if err != nil {
				return errors.InternalServerError("store.Blob.Write", err.Error())
			}
		} else {
			// subsequent message recieved from the stream
			buf.Write(req.Blob)
		}
	}

	// ensure the blob was sent over the stream
	if buf == nil {
		return errors.BadRequest("store.Blob.Write", "No blob was sent")
	}

	// execute the request
	err := store.DefaultBlobStore.Write(key, buf, gostore.BlobNamespace(options.Namespace))
	if err == gostore.ErrMissingKey {
		return errors.BadRequest("store.Blob.Write", "Missing key")
	} else if err != nil {
		return errors.InternalServerError("store.Blob.Write", err.Error())
	}

	// close the stream
	return stream.SendAndClose(&pb.BlobWriteResponse{})
}

func (b *blobHandler) Delete(ctx context.Context, req *pb.BlobDeleteRequest, rsp *pb.BlobDeleteResponse) error {
	// parse the options
	if ns := req.GetOptions().GetNamespace(); len(ns) == 0 {
		req.Options = &pb.BlobOptions{
			Namespace: namespace.FromContext(ctx),
		}
	}

	// authorize the request
	if err := authns.Authorize(ctx, req.Options.Namespace); err == authns.ErrForbidden {
		return errors.Forbidden("store.Blob.Delete", err.Error())
	} else if err == authns.ErrUnauthorized {
		return errors.Unauthorized("store.Blob.Delete", err.Error())
	} else if err != nil {
		return errors.InternalServerError("store.Blob.Delete", err.Error())
	}

	// execute the request
	err := store.DefaultBlobStore.Delete(req.Key, gostore.BlobNamespace(req.Options.Namespace))
	if err == gostore.ErrNotFound {
		return errors.NotFound("store.Blob.Delete", "Blob not found")
	} else if err == gostore.ErrMissingKey {
		return errors.BadRequest("store.Blob.Delete", "Missing key")
	} else if err != nil {
		return errors.InternalServerError("store.Blob.Delete", err.Error())
	}

	return nil
}
