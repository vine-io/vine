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

package runtime

const (
	// EventTopic the events are published to
	EventTopic = "runtime"

	// EventServiceCreated is the topic events are published to when a service is created
	EventServiceCreated = "service.created"
	// EventServiceUpdated is the topic events are published to when a service is updated
	EventServiceUpdated = "service.updated"
	// EventServiceDeleted is the topic events are published to when a service is deleted
	EventServiceDeleted       = "service.deleted"
	EventNamespaceCreated     = "namespace.created"
	EventNamespaceDeleted     = "namespace.deleted"
	EventNetworkPolicyCreated = "networkpolicy.created"
	EventNetworkPolicyUpdated = "networkpolicy.updated"
	EventNetworkPolicyDeleted = "networkpolicy.deleted"
	EventResourceQuotaCreated = "resourcequota.created"
	EventResourceQuotaUpdated = "resourcequota.updated"
	EventResourceQuotaDeleted = "resourcequota.deleted"
)

// EventPayload which is published with runtime events
type EventPayload struct {
	Type      string
	Service   *Service
	Namespace string
}

// EventResourcePayload which is published with runtime resource events
type EventResourcePayload struct {
	Type          string
	Name          string
	Namespace     string
	NetworkPolicy *NetworkPolicy
	ResourceQuota *ResourceQuota
	Service       *Service
}
