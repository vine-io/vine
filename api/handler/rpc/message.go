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

package rpc

type Message struct {
	data []byte
}

func (m *Message) ProtoMessage() {}

func (m *Message) Reset() {
	*m = Message{}
}

func (m *Message) String() string {
	return string(m.data)
}

func (m *Message) Marshal() ([]byte, error) {
	return m.data, nil
}

func (m *Message) Unmarshal(data []byte) error {
	m.data = data
	return nil
}

func NewMessage(data []byte) *Message {
	return &Message{data}
}
