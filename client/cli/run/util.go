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

package runtime

import (
	"io"

	"github.com/lack-io/cli"

	pb "github.com/lack-io/vine/proto/runtime"
	"github.com/lack-io/vine/service/build/util/tar"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/context"
	"github.com/lack-io/vine/service/runtime"
	"github.com/lack-io/vine/service/runtime/source/git"
)

const bufferSize = 1024

// upload source to the server. will return the source id, e.g. source://foo-bar and an error if
// one occured. The ID returned can be used as a source in runtime.Create.
func upload(ctx *cli.Context, srv *runtime.Service, source *git.Source) (string, error) {
	// if the source exists within a local git repository, archive the whole repository, otherwise
	// just archive the folder
	var data io.Reader
	var err error
	if len(source.LocalRepoRoot) > 0 {
		data, err = tar.Archive(source.LocalRepoRoot)
	} else {
		data, err = tar.Archive(source.FullPath)
	}
	if err != nil {
		return "", err
	}

	// create an upload stream
	cli := pb.NewSourceService("runtime", client.DefaultClient)
	stream, err := cli.Upload(context.DefaultContext, client.WithAuthToken())
	if err != nil {
		return "", err
	}

	// read bytes from the tar and stream it to the server
	buffer := make([]byte, bufferSize)
	var sentService bool
	for {
		num, err := data.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		req := &pb.UploadRequest{Data: buffer[:num]}

		// construct the service object, we'll send this on the first message only to reduce the amount of
		// data needed to be streamed
		if !sentService {
			req.Service = &pb.Service{Name: srv.Name, Version: srv.Version}
			sentService = true
		}

		if err := stream.Send(req); err != nil {
			return "", err
		}
	}

	// wait for the server to process the source
	rsp, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}
	return rsp.Id, nil
}
