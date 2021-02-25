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

package memory

import (
	"encoding/json"
	"testing"

	regpb "github.com/lack-io/vine/proto/apis/registry"
)

func TestWatcher(t *testing.T) {
	w := &Watcher{
		id:   "test",
		res:  make(chan *regpb.Result),
		exit: make(chan bool),
	}

	go func() {
		w.res <- &regpb.Result{}
	}()

	_, err := w.Next()
	if err != nil {
		t.Fatal("unexpected err", err)
	}

	w.Stop()

	if _, err := w.Next(); err == nil {
		t.Fatal("expected error on Next()")
	}
}

type A struct {
	Ref             string `json:"$ref"`
	ApplicationJson string `json:"application/json"`
	N               string `json:"n"`
}

func TestTodo(t *testing.T) {
	a := &A{
		Ref:             "http://example.org",
		ApplicationJson: "aa",
		N:               "'200'",
	}
	v, _ := json.Marshal(a)
	t.Log(string(v))
}
