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
