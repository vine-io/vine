// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"net/http"

	json "github.com/json-iterator/go"

	proto "github.com/lack-io/vine/proto/config"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/config"
	"github.com/lack-io/vine/service/context"
	"github.com/lack-io/vine/service/errors"
)

var (
	defaultNamespace = "vine"
	name             = "config"
)

type srv struct {
	opts      config.Options
	namespace string
	client    proto.ConfigService
}

func (m *srv) Get(path string, options ...config.Option) (config.Value, error) {
	o := config.Options{}
	for _, option := range options {
		option(&o)
	}
	nullValue := config.NewJSONValue([]byte("null"))
	req, err := m.client.Get(context.DefaultContext, &proto.GetRequest{
		Namespace: m.namespace,
		Path:      path,
		Options: &proto.Options{
			Secret: o.Secret,
		},
	}, client.WithAuthToken())
	if verr := errors.FromErr(err); verr != nil && verr.Code == http.StatusNotFound {
		return nullValue, nil
	} else if err != nil {
		return nullValue, err
	}

	return config.NewJSONValue([]byte(req.Value.Data)), nil
}

func (m *srv) Set(path string, value interface{}, options ...config.Option) error {
	o := config.Options{}
	for _, option := range options {
		option(&o)
	}
	dat, _ := json.Marshal(value)
	_, err := m.client.Set(context.DefaultContext, &proto.SetRequest{
		Namespace: m.namespace,
		Path:      path,
		Value: &proto.Value{
			Data: string(dat),
		},
		Options: &proto.Options{
			Secret: o.Secret,
		},
	}, client.WithAuthToken())
	return err
}

func (m *srv) Delete(path string, options ...config.Option) error {
	_, err := m.client.Delete(context.DefaultContext, &proto.DeleteRequest{
		Namespace: m.namespace,
		Path:      path,
	}, client.WithAuthToken())
	return err
}

func (m *srv) String() string {
	return "service"
}

func NewConfig(namespace string) *srv {
	addr := name
	if len(namespace) == 0 {
		namespace = defaultNamespace
	}

	s := &srv{
		namespace: namespace,
		client:    proto.NewConfigService(addr, client.DefaultClient),
	}

	return s
}
