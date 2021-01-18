// Copyright 2020 lack
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

package service

import (
	"context"

	"github.com/lack-io/vine/service/client"
)

type event struct {
	c     client.Client
	topic string
}

func (e *event) Publish(ctx context.Context, msg interface{}, opts ...client.PublishOption) error {
	return e.c.Publish(ctx, e.c.NewMessage(e.topic, msg), opts...)
}
