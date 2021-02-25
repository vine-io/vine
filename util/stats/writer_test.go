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

package stats

import (
	"net/http"
	"testing"
)

type testResponseWriter struct {
	code int
}

func (w *testResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (w *testResponseWriter) WriteHeader(c int) {
	w.code = c
}

func TestWriter(t *testing.T) {
	tw := &testResponseWriter{}
	w := &writer{tw, 0}
	w.WriteHeader(200)

	if w.status != 200 {
		t.Fatalf("Expected status 200 got %d", w.status)
	}
	if tw.code != 200 {
		t.Fatalf("Expected status 200 got %d", tw.code)
	}
}
