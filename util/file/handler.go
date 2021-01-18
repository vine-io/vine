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

package file

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lack-io/vine/proto/errors"
	proto "github.com/lack-io/vine/proto/file"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/server"
)

// NewHandler is a handler that can be registered with a vine Server
func NewHandler(readDir string) proto.FileHandler {
	return &handler{
		readDir: readDir,
		session: &session{
			files: make(map[int64]*os.File),
		},
	}
}

// RegisterHandler is a convenience method for registering a handler
func RegisterHandler(s server.Server, readDir string) {
	proto.RegisterFileHandler(s, NewHandler(readDir))
}

type handler struct {
	readDir string
	session *session
}

func (h *handler) Open(ctx context.Context, req *proto.OpenRequest, rsp *proto.OpenResponse) error {
	path := filepath.Join(h.readDir, req.Filename)
	flags := os.O_CREATE | os.O_RDWR
	if req.GetTruncate() {
		flags = flags | os.O_TRUNC
	}
	file, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return errors.InternalServerError("go.vine.server", err.Error())
	}

	rsp.Id = h.session.Add(file)
	rsp.Result = true

	log.Debugf("Open %s, sessionId=%d", req.Filename, rsp.Id)

	return nil
}

func (h *handler) Close(ctx context.Context, req *proto.CloseRequest, rsp *proto.CloseResponse) error {
	h.session.Delete(req.Id)
	log.Debugf("Close sessionId=%d", req.Id)
	return nil
}

func (h *handler) Stat(ctx context.Context, req *proto.StatRequest, rsp *proto.StatResponse) error {
	path := filepath.Join(h.readDir, req.Filename)
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.InternalServerError("go.vine.srv.file", err.Error())
	}

	if fi.IsDir() {
		rsp.Type = "Directory"
	} else {
		rsp.Type = "File"
		rsp.Size_ = fi.Size()
	}

	rsp.LastModified = fi.ModTime().Unix()
	log.Debugf("Stat %s, %#v", req.Filename, rsp)

	return nil
}

func (h *handler) Read(ctx context.Context, req *proto.ReadRequest, rsp *proto.ReadResponse) error {
	file := h.session.Get(req.Id)
	if file == nil {
		return errors.InternalServerError("go.vine.srv.file", "You must call open first.")
	}

	rsp.Data = make([]byte, req.Size())
	n, err := file.ReadAt(rsp.Data, req.Offset)
	if err != nil && err != io.EOF {
		return errors.InternalServerError("go.vine.srv.file", err.Error())
	}

	if err == io.EOF {
		rsp.Eof = true
	}

	rsp.Size_ = int64(n)
	rsp.Data = rsp.Data[:n]

	log.Debugf("Read sessionId=%d, Offset=%d, n=%d", req.Id, req.Offset, rsp.Size)

	return nil
}

func (h *handler) Write(ctx context.Context, req *proto.WriteRequest, rsp *proto.WriteResponse) error {
	file := h.session.Get(req.Id)
	if file == nil {
		return errors.InternalServerError("go.vine.srv.file", "You must call open first.")
	}

	if _, err := file.WriteAt(req.GetData(), req.GetOffset()); err != nil {
		return err
	}

	log.Debugf("Write sessionId=%d, Offset=%d, n=%d", req.Id, req.Offset)

	return nil
}

type session struct {
	sync.Mutex
	files   map[int64]*os.File
	counter int64
}

func (s *session) Add(file *os.File) int64 {
	s.Lock()
	defer s.Unlock()

	s.counter += 1
	s.files[s.counter] = file

	return s.counter
}

func (s *session) Get(id int64) *os.File {
	s.Lock()
	defer s.Unlock()
	return s.files[id]
}

func (s *session) Delete(id int64) {
	s.Lock()
	defer s.Unlock()

	if file, exist := s.files[id]; exist {
		file.Close()
		delete(s.files, id)
	}
}

func (s *session) Len() int {
	return len(s.files)
}
