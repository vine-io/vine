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

package errors

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gogo/protobuf/proto"
	json "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Id       string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Code     int32    `protobuf:"varint,2,opt,name=code,proto3" json:"code,omitempty"`
	Detail   string   `protobuf:"bytes,3,opt,name=detail,proto3" json:"detail,omitempty"`
	Status   string   `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
	Position string   `protobuf:"bytes,5,opt,name=position,proto3" json:"position,omitempty"`
	Child    *Child   `protobuf:"bytes,6,opt,name=child,proto3" json:"child,omitempty"`
	Stacks   []*Stack `protobuf:"bytes,7,rep,name=stacks,proto3" json:"stacks,omitempty"`
}

func (e *Error) Reset()         { *e = Error{} }
func (e *Error) String() string { return proto.CompactTextString(e) }
func (*Error) ProtoMessage()    {}

type Child struct {
	Code   int32  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Detail string `protobuf:"bytes,2,opt,name=detail,proto3" json:"detail,omitempty"`
}

func (m *Child) Reset()         { *m = Child{} }
func (m *Child) String() string { return proto.CompactTextString(m) }
func (*Child) ProtoMessage()    {}

type Stack struct {
	Code     int32  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Detail   string `protobuf:"bytes,2,opt,name=detail,proto3" json:"detail,omitempty"`
	Position string `protobuf:"bytes,3,opt,name=position,proto3" json:"position,omitempty"`
}

func (m *Stack) Reset()         { *m = Stack{} }
func (m *Stack) String() string { return proto.CompactTextString(m) }
func (*Stack) ProtoMessage()    {}

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

func (e *Error) WithId(id string) *Error {
	e.Id = id
	return e
}

func (e *Error) WithCode(code int32) *Error {
	e.Code = code
	e.Status = http.StatusText(int(code))
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

// PreconditionFailed generates a 412 error.
func PreconditionFailed(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 412)
}

// TooManyRequests generates a 429 error.
func TooManyRequests(id, format string, a ...interface{}) *Error {
	return New(id, fmt.Sprintf(format, a...), 429)
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

	if err != nil {
		if se, ok := err.(interface {
			GRPCStatus() *status.Status
		}); ok {
			s := se.GRPCStatus()
			switch s.Code() {
			case codes.OK:
				return &Error{Code: 0}
			case codes.Canceled:
				return Timeout("", s.Message())
			case codes.DeadlineExceeded:
				return GatewayTimeout("", s.Message())
			case codes.NotFound:
				return NotFound("", s.Message())
			case codes.AlreadyExists:
				return Conflict("", s.Message())
			case codes.PermissionDenied:
				return Forbidden("", s.Message())
			case codes.ResourceExhausted:
				return TooManyRequests("", s.Message())
			case codes.FailedPrecondition:
				return PreconditionFailed("", s.Message())
			case codes.Aborted:
				return Conflict("", s.Message())
			case codes.OutOfRange:
				return BadGateway("", s.Message())
			case codes.Unimplemented:
				return NotImplemented("", s.Message())
			case codes.Internal:
				return InternalServerError("", s.Message())
			case codes.Unavailable:
				return ServiceUnavailable("", s.Message())
			case codes.DataLoss:
				return InternalServerError("", s.Message())
			case codes.Unauthenticated:
				return Unauthorized("", s.Message())
			}
			return Parse(se.GRPCStatus().Message())
		}
	}

	return Parse(err.Error())
}
