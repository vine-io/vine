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

	pb "github.com/lack-io/vine/proto/usage"
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
