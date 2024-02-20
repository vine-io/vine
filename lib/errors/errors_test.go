// MIT License
//
// Copyright (c) 2020 The vine Authors
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
	er "errors"
	"fmt"
	"net/http"
	"testing"
)

func TestFromError(t *testing.T) {
	err := NotFound("go.vine.test", "%s", "example")
	merr := FromErr(err)
	if merr.Id != "go.vine.test" || merr.Code != 404 {
		t.Fatalf("invalid conversation %v != %v", err, merr)
	}
	err1 := er.New(err.Error())
	merr = FromErr(err1)
	if merr.Id != "go.vine.test" || merr.Code != 404 {
		t.Fatalf("invalid conversation %v != %v", err, merr)
	}

}

func TestEqual(t *testing.T) {
	err1 := NotFound("myid1", "msg1")
	err2 := NotFound("myid2", "msg2")

	if !Equal(err1, err2) {
		t.Fatal("errors must be equal")
	}

	err3 := er.New("my test err")
	if Equal(err1, err3) {
		t.Fatal("errors must be not equal")
	}

}

func TestErrors(t *testing.T) {
	testData := []*Error{
		{
			Id:     "test",
			Code:   500,
			Detail: "Internal server error",
			Status: http.StatusText(500),
		},
	}

	for _, e := range testData {
		ne := New(e.Id, e.Detail, e.Code)

		if e.Error() != ne.Error() {
			t.Fatalf("Expected %s got %s", e.Error(), ne.Error())
		}

		pe := Parse(ne.Error())

		if pe == nil {
			t.Fatalf("Expected error got nil %v", pe)
		}

		if pe.Id != e.Id {
			t.Fatalf("Expected %s got %s", e.Id, pe.Id)
		}

		if pe.Detail != e.Detail {
			t.Fatalf("Expected %s got %s", e.Detail, pe.Detail)
		}

		if pe.Code != e.Code {
			t.Fatalf("Expected %d got %d", e.Code, pe.Code)
		}

		if pe.Status != e.Status {
			t.Fatalf("Expected %s got %s", e.Status, pe.Status)
		}
	}
}

func TestError_WithPos(t *testing.T) {
	e := BadRequest("go.vine.client", "test").WithPos()
	t.Log(e)
}

func TestError_WithChild(t *testing.T) {
	e := BadRequest("go.vine.client", "test").WithChild(111, "%v", fmt.Errorf("err1"))
	t.Log(e)
}

func TestError_WithStack(t *testing.T) {
	e := New("go.vine.client", "test stack", 400)
	if len(e.Stacks) != 0 {
		t.Fatalf("Stack = %v", e.Stacks)
	}

	e.WithStack(10001, "stack information", true)
	if len(e.Stacks) == 0 {
		t.Fatalf("call Stack() failed")
	}

	t.Log(e)
}
