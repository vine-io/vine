package encoding

import (
	"github.com/vine-io/vine/core/codec"
	"github.com/vine-io/vine/core/codec/bytes"
	"github.com/vine-io/vine/core/codec/json"
	"github.com/vine-io/vine/core/codec/proto"
	"github.com/vine-io/vine/core/codec/yaml"
)

func init() {
	RegisterMarshaler("bytes", &bytes.Marshaler{})
	RegisterMarshaler("json", &json.Marshaler{})
	RegisterMarshaler("yaml", &yaml.Marshaler{})
	RegisterMarshaler("proto", &proto.Marshaler{})
}

// mSet the set of Marshaler
var mSet = map[string]codec.Marshaler{}

// RegisterMarshaler puts the implement of Marshaler to mSet
func RegisterMarshaler(name string, m codec.Marshaler) {
	mSet[name] = m
}

// GetMarshaler gets implement from mSet
func GetMarshaler(name string) (codec.Marshaler, bool) {
	m, ok := mSet[name]
	return m, ok
}
