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

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
)

// protoc -I. -I$GOPATH/src --gogofast_out=plugins=grpc:. --vine_out=:. errors.proto

// New generates a custom error.
func New(id, detail string, code int32) *Error {
	e := &Error{
		Id:     id,
		Code:   code,
		Detail: detail,
		Status: http.StatusText(int(code)),
	}
	return e
}

// WithChild fills Error.Child
func (e *Error) WithChild(code int32, format string, a ...interface{}) *Error {
	e.Child = &Child{
		Code:   code,
		Detail: fmt.Sprintf(format, a...),
	}
	return e
}

// WithPos fills Error.Position
func (e *Error) WithPos() *Error {
	_, file, line, _ := runtime.Caller(1)
	if index := strings.Index(file, "/src/"); index != -1 {
		file = file[index+5:]
	}
	file = strings.Replace(file, string(filepath.Separator), "/", -1)
	e.Position = fmt.Sprintf("%s:%d", file, line)
	return e
}

// WithStack push stack information to Error
func (e *Error) WithStack(code int32, detail string, pos ...bool) *Error {
	if e.Stacks == nil {
		e.Stacks = make([]*Stack, 0)
	}
	stack := &Stack{Code: code, Detail: detail}
	if pos != nil && pos[0] {
		_, file, line, _ := runtime.Caller(1)
		if index := strings.Index(file, "/src/"); index != -1 {
			file = file[index+5:]
		}
		file = strings.Replace(file, string(filepath.Separator), "/", -1)
		stack.Position = fmt.Sprintf("%s:%d", file, line)
	}
	e.Stacks = append(e.Stacks, stack)
	return e
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// Parse tries to parse a JSON string into an error. If that
// fails, it will set the given string as the error detail.
func Parse(err string) *Error {
	e := new(Error)
	errr := json.Unmarshal([]byte(err), e)
	if errr != nil {
		e.Detail = err
	}
	return e
}

// BadRequest generates a 400 error.
func BadRequest(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 400)
}

// Unauthorized generates a 401 error.
func Unauthorized(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 401)
}

// Forbidden generates a 403 error.
func Forbidden(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 403)
}

// NotFound generates a 404 error.
func NotFound(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 404)
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 405)
}

// Timeout generates a 408 error.
func Timeout(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 408)
}

// Conflict generates a 409 error.
func Conflict(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 409)
}

// InternalServerError generates a 500 error.
func InternalServerError(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 500)
}

// NotImplemented generates a 501 error
func NotImplemented(id, format string, a ...interface{}) *Error {
	return &Error{
		Id:     id,
		Code:   501,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(501),
	}
}

// BadGateway generates a 502 error
func BadGateway(id, format string, a ...interface{}) *Error {
	return &Error{
		Id:     id,
		Code:   502,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(502),
	}
}

// ServiceUnavailable generates a 503 error
func ServiceUnavailable(id, format string, a ...interface{}) *Error {
	return &Error{
		Id:     id,
		Code:   503,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(503),
	}
}

// GatewayTimeout generates a 504 error
func GatewayTimeout(id, format string, a ...interface{}) *Error {
	return &Error{
		Id:     id,
		Code:   504,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(504),
	}
}

// Equal tries to compare errors
func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return err1 == err2
	}

	if verr1.Code != verr2.Code {
		return false
	}

	return true
}

// FromErr try to convert go error go *Error
func FromErr(err error) *Error {
	if verr, ok := err.(*Error); ok && verr != nil {
		return verr
	}

	return Parse(err.Error())
}
