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

package client

import (
	"context"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	results := []time.Duration{
		0 * time.Second,
		100 * time.Millisecond,
		600 * time.Millisecond,
		1900 * time.Millisecond,
		4300 * time.Millisecond,
		7900 * time.Millisecond,
	}

	c := NewClient()

	for i := 0; i < 5; i++ {
		d, err := exponentialBackoff(context.TODO(), c.NewRequest("test", "test", nil), i)
		if err != nil {
			t.Fatal(err)
		}

		if d != results[i] {
			t.Fatalf("Expected equal than %v, got %v", results[i], d)
		}
	}
}
