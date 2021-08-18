// MIT License
//
// Copyright (c) 2021 Lack
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

package http

import "github.com/vine-io/vine/core/codec"

type httpMessage struct {
	topic string
	payload interface{}
	contentType string
	header map[string]string
	body []byte
	codec codec.Reader
}

func (r *httpMessage) Topic() string {
	return r.topic
}

func (r *httpMessage) Payload() interface{} {
	return r.payload
}

func (r *httpMessage) ContentType() string {
	return r.contentType
}

func (r *httpMessage) Header() map[string]string {
	return r.header
}

func (r *httpMessage) Body() []byte {
	return r.body
}

func (r *httpMessage) Codec() codec.Reader {
	return r.codec
}