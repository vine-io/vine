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

package handler

import (
	"context"
	"fmt"
	"io"

	pb "github.com/lack-io/vine/proto/runtime"
	"github.com/lack-io/vine/service/auth"
	"github.com/lack-io/vine/service/errors"
	"github.com/lack-io/vine/service/store"
	gostore "github.com/lack-io/vine/service/store"
)

const bufferSize = 1024

// Build implements the proto build service interface
type Build struct{}

func (b *Build) Read(ctx context.Context, req *pb.Service, stream pb.Build_ReadStream) error {
	defer stream.Close()

	// authorize the request
	acc, ok := auth.AccountFromContext(ctx)
	if !ok {
		return errors.Unauthorized("runtime.Build.Read", "An account is required to read builds")
	}

	// validate the request
	if len(req.Name) == 0 {
		return errors.BadRequest("runtime.Build.Read", "Missing name")
	}
	if len(req.Version) == 0 {
		return errors.BadRequest("runtime.Build.Read", "Missing version")
	}

	// lookup the build from the blob store
	key := fmt.Sprintf("build://%v:%v", req.Name, req.Version)
	build, err := store.DefaultBlobStore.Read(key, gostore.BlobNamespace(acc.Issuer))
	if err == store.ErrNotFound {
		return errors.NotFound("runtime.Build.Read", "Build not found")
	} else if err != nil {
		return err
	}

	// read bytes from the store and stream it to the client
	buffer := make([]byte, bufferSize)
	for {
		num, err := build.Read(buffer)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return errors.InternalServerError("runtime.Build.Read", "Error reading build from store: %v", err)
		}

		if err := stream.Send(&pb.BuildReadResponse{Data: buffer[:num]}); err != nil {
			return err
		}
	}
}
