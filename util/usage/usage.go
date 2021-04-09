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

// Package usage tracks vine usage
package usage

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"

	pb "github.com/lack-io/vine/proto/apis/usage"
	"github.com/lack-io/vine/util/version"
)

var (
	// usage url
	u = "https://go.vine.mu/usage"
	// usage agent
	a = "vine/usage"
	// usage version
	v = version.V
	// 24 hour window
	w = 8.64e13
)

// New generates a new usage report to be filled in
func New(service string) *pb.Usage {
	id := fmt.Sprintf("vine.%s.%s.%s", service, version.V, uuid.New().String())
	svc := "vine." + service

	if len(service) == 0 {
		id = fmt.Sprintf("vine.%s.%s", version.V, uuid.New().String())
		svc = "vine"
	}

	sum := sha256.Sum256([]byte(id))

	return &pb.Usage{
		Service:   svc,
		Version:   v,
		Id:        fmt.Sprintf("%x", sum),
		Timestamp: uint64(time.Now().UnixNano()),
		Window:    uint64(w),
		Metrics: &pb.Metrics{
			Count: make(map[string]uint64),
		},
	}
}

// Report reports the current usage
func Report(ug *pb.Usage) error {
	if v := os.Getenv("VINE_REPORT_USAGE"); v == "false" {
		return nil
	}

	// update timestamp/window
	now := uint64(time.Now().UnixNano())
	ug.Window = now - ug.Timestamp
	ug.Timestamp = now

	p, err := proto.Marshal(ug)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u, bytes.NewReader(p))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("User-Agent", a)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	io.Copy(ioutil.Discard, rsp.Body)
	return nil
}
