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

package file

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lack-io/vine/core/server"
	log "github.com/lack-io/vine/lib/logger"
	"github.com/lack-io/vine/proto/apis/errors"
	proto "github.com/lack-io/vine/proto/services/file"
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
		return errors.InternalServerError("go.vine.svc.file", err.Error())
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
		return errors.InternalServerError("go.vine.svc.file", "You must call open first.")
	}

	rsp.Data = make([]byte, req.Size())
	n, err := file.ReadAt(rsp.Data, req.Offset)
	if err != nil && err != io.EOF {
		return errors.InternalServerError("go.vine.svc.file", err.Error())
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
		return errors.InternalServerError("go.vine.svc.file", "You must call open first.")
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
