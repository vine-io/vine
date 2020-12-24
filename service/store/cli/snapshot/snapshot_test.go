// Copyright 2020 The vine Authors
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

package snapshot

import (
	"testing"
	"time"

	"github.com/lack-io/vine/service/store"
)

func TestFileSnapshot(t *testing.T) {
	f := NewFileSnapshot(Destination("invalid"))
	if err := f.Init(); err == nil {
		t.Error(err)
	}
	if err := f.Init(Destination("file:///tmp/test-snapshot")); err != nil {
		t.Error(err)
	}

	recordChan, err := f.Start()
	if err != nil {
		t.Error(err)
	}

	for _, td := range testData {
		recordChan <- td
	}
	close(recordChan)
	f.Wait()

	r := NewFileRestore(Source("invalid"))
	if err := r.Init(); err == nil {
		t.Error(err)
	}
	if err := r.Init(Source("file:///tmp/test-snapshot")); err != nil {
		t.Error(err)
	}

	returnChan, err := r.Start()
	if err != nil {
		t.Error(err)
	}
	var receivedData []*store.Record
	for r := range returnChan {
		println("decoded", r.Key)
		receivedData = append(receivedData, r)
	}

}

var testData = []*store.Record{
	{
		Key:    "foo",
		Value:  []byte(`foo`),
		Expiry: time.Until(time.Now().Add(5 * time.Second)),
	},
	{
		Key:    "bar",
		Value:  []byte(`bar`),
		Expiry: time.Until(time.Now().Add(5 * time.Second)),
	},
	{
		Key:    "baz",
		Value:  []byte(`baz`),
		Expiry: time.Until(time.Now().Add(5 * time.Second)),
	},
}
