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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lack-io/vine/service/config"
)

func TestConfig(t *testing.T) {
	data := []byte(`{"foo": "bar"}`)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		fh.Close()
		os.Remove(path)
	}()
	_, err = fh.Write(data)
	if err != nil {
		t.Error(err)
	}

	conf, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	conf.Load(NewSource(WithPath(path)))
	// simulate multiple close
	go conf.Close()
	go conf.Close()
}

func TestFile(t *testing.T) {
	data := []byte(`{"foo": "bar"}`)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	_, err = fh.Write(data)
	if err != nil {
		t.Error(err)
	}

	f := NewSource(WithPath(path))
	c, err := f.Read()
	if err != nil {
		t.Error(err)
	}
	if string(c.Data) != string(data) {
		t.Logf("%+v", c)
		t.Error("data from file does not match")
	}
}
