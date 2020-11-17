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

package codes

import (
	"fmt"
	"runtime"
	"strings"

	metav1 "github.com/lack-io/vine/internal/meta/v1"

	json "github.com/json-iterator/go"
)

var (
	project = ""
	service = ""
)

// Load will be called when vine component initialized. Then project and service
// will be assigned.
func Load(pro, sv string) {
	project, service = pro, sv
}

// Codes contains *metav1.Status. It can be mutually converted with *metev1.Status.
type Codes struct {
	s *metav1.Status
}

// New create an Codes by Codez
func New(codez Codez) *Codes {
	return &Codes{
		s: &metav1.Status{
			Project: project,
			Service: service,
			Code:    codez.code,
			Message: codez.msg,
		},
	}
}

// FromStatus create a Codes by *metav1.Status
func FromStatus(status *metav1.Status) *Codes {
	return &Codes{s: status}
}

// Convert create Codes by error. If err is nil, calls IsOK will return true.
func Convert(err error) *Codes {
	if err == nil {
		return New(OK)
	}
	if v, ok := err.(interface {
		ToStatus() *metav1.Status
	}); ok {
		return FromStatus(v.ToStatus())
	}
	msg := err.Error()
	status := &metav1.Status{}
	if e := json.Unmarshal([]byte(msg), &status); e != nil {
		return &Codes{s: status}
	}
	return New(Unknown)
}

// Codez returns Codez by Codes
func (c *Codes) Codez() Codez {
	return Codez{code: c.s.Code, msg: c.s.Message}
}

// Wrap modifies Desc field. If Desc field is empty, instead err, otherwise
// append err string.
func (c *Codes) Wrap(err error) *Codes {
	if err != nil {
		return c
	}
	if c.s.Desc == "" {
		c.s.Desc = err.Error()
	} else {
		c.s.Desc = fmt.Sprintf("%s: %v", c.s.Desc, err)
	}
	return c
}

// WrapDesc modifies Desc field to text
func (c *Codes) WrapDesc(text string) *Codes {
	c.s.Desc = text
	return c
}

// Call calls CallDepth method, d = 1
func (c *Codes) Call() *Codes {
	return c.CallDepth(1)
}

// CallDepth calls runtime.Caller to collects error position
func (c *Codes) CallDepth(d int) *Codes {
	_, file, line, _ := runtime.Caller(d+1)
	if i := strings.Index(file, "src"); i != -1 {
		file = file[i+4:]
	}
	c.s.Pos = fmt.Sprintf("%s:%d", file, line)
	return c
}

// Stack collects the context of the error
func (c *Codes) Stack(msg string) *Codes {
	if c.s.Details == nil {
		c.s.Details = []metav1.StatusDetail{}
	}
	_, file, line, _ := runtime.Caller(1)
	if i := strings.Index(file, "src"); i != -1 {
		file = file[i+4:]
	}
	c.s.Details = append(c.s.Details, metav1.StatusDetail{
		Pos:  fmt.Sprintf("%s:%d", file, line),
		Desc: msg,
	})
	return c
}

// ToStatus returns inner *metav1.Status
func (c *Codes) ToStatus() *metav1.Status {
	return c.s
}

// Err convert Codes to error. If IsOK is true, error will be nil
func (c *Codes) Err() error {
	if c.IsOK() {
		return nil
	}
	s, _ := json.MarshalToString(c.s)
	return fmt.Errorf("%s", s)
}

// IsOK return true if Codes is nil or Code is 0
func (c *Codes) IsOK() bool {
	if c.s == nil {
		return true
	}
	return c.s.Code == OK.code
}

func (c *Codes) String() string {
	return c.s.String()
}
