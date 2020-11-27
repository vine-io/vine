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

package client

type message struct {
	topic       string
	contentType string
	payload     interface{}
}

func newMessage(topic string, payload interface{}, contentType string, opts ...MessageOption) Message {
	var options MessageOptions
	for _, o := range opts {
		o(&options)
	}

	if len(options.ContentType) > 0 {
		contentType = options.ContentType
	}

	return &message{
		payload:     payload,
		topic:       topic,
		contentType: contentType,
	}
}

func (m *message) ContentType() string {
	return m.contentType
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() interface{} {
	return m.payload
}
