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

package web

import (
	regpb "github.com/lack-io/vine/proto/apis/registry"
)

type sortedServices struct {
	services []*regpb.Service
}

func (s sortedServices) Len() int {
	return len(s.services)
}

func (s sortedServices) Less(i, j int) bool {
	return s.services[i].Name < s.services[j].Name
}

func (s sortedServices) Swap(i, j int) {
	s.services[i], s.services[j] = s.services[j], s.services[i]
}
