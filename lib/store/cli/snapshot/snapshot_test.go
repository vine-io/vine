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

package snapshot

import (
	"testing"
	"time"

	"github.com/lack-io/vine/lib/store"
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
