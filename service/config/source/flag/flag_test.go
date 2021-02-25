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

package flag

import (
	"encoding/json"
	"flag"
	"testing"
)

var (
	dbuser = flag.String("database-user", "default", "db user")
	dbhost = flag.String("database-host", "", "db host")
	dbpw   = flag.String("database-password", "", "db pw")
)

func initTestFlags() {
	flag.Set("database-host", "localhost")
	flag.Set("database-password", "some-password")
	flag.Parse()
}

func TestFlagsrc_Read(t *testing.T) {
	initTestFlags()
	source := NewSource()
	c, err := source.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data, &actual); err != nil {
		t.Error(err)
	}

	actualDB := actual["database"].(map[string]interface{})
	if actualDB["host"] != *dbhost {
		t.Errorf("expected %v got %v", *dbhost, actualDB["host"])
	}

	if actualDB["password"] != *dbpw {
		t.Errorf("expected %v got %v", *dbpw, actualDB["password"])
	}

	// unset flags should not be loaded
	if actualDB["user"] != nil {
		t.Errorf("expected %v got %v", nil, actualDB["user"])
	}
}

func TestFlagsrc_ReadAll(t *testing.T) {
	initTestFlags()
	source := NewSource(IncludeUnset(true))
	c, err := source.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data, &actual); err != nil {
		t.Error(err)
	}

	actualDB := actual["database"].(map[string]interface{})

	// unset flag defaults should be loaded
	if actualDB["user"] != *dbuser {
		t.Errorf("expected %v got %v", *dbuser, actualDB["user"])
	}
}
