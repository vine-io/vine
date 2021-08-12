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

package json

import (
	"errors"
	"time"

	"github.com/imdario/mergo"

	"github.com/vine-io/vine/lib/config/encoder"
	"github.com/vine-io/vine/lib/config/encoder/json"
	"github.com/vine-io/vine/lib/config/reader"
	"github.com/vine-io/vine/lib/config/source"
)

type jsonReader struct {
	opts reader.Options
	json encoder.Encoder
}

func (j *jsonReader) Merge(changes ...*source.ChangeSet) (*source.ChangeSet, error) {
	var merged map[string]interface{}

	for _, m := range changes {
		if m == nil {
			continue
		}

		if len(m.Data) == 0 {
			continue
		}

		codec, ok := j.opts.Encoding[m.Format]
		if !ok {
			// fallback
			codec = j.json
		}

		var data map[string]interface{}
		if err := codec.Decode(m.Data, &data); err != nil {
			return nil, err
		}
		if err := mergo.Map(&merged, data, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	b, err := j.json.Encode(merged)
	if err != nil {
		return nil, err
	}

	cs := &source.ChangeSet{
		Timestamp: time.Now(),
		Data:      b,
		Source:    "json",
		Format:    j.json.String(),
	}
	cs.CheckSum = cs.Sum()

	return cs, nil
}

func (j *jsonReader) Values(ch *source.ChangeSet) (reader.Values, error) {
	if ch == nil {
		return nil, errors.New("changeset is nil")
	}
	if ch.Format != "json" {
		return nil, errors.New("unsupported format")
	}
	return newValues(ch)
}

func (j *jsonReader) String() string {
	return "json"
}

// NewReader creates a json reader
func NewReader(opts ...reader.Option) reader.Reader {
	options := reader.NewOptions(opts...)
	return &jsonReader{
		json: json.NewEncoder(),
		opts: options,
	}
}
