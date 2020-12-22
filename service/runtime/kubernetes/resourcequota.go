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

package kubernetes

import (
	"github.com/lack-io/vine/internal/kubernetes/client"
	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
)

// createResourceQuota creates a resourcequota resource
func (k *kubernetes) createResourceQuota(resourceQuota *runtime.ResourceQuota) error {
	err := k.client.Create(&client.Resource{
		Kind:  "resourcequota",
		Value: client.NewResourceQuota(resourceQuota),
	}, client.CreateNamespace(resourceQuota.Namespace))
	if err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("Error creating resource %s: %v", resourceQuota.String(), err)
		}
	}
	return err
}

// updateResourceQuota updates a resourcequota resource in-place
func (k *kubernetes) updateResourceQuota(resourceQuota *runtime.ResourceQuota) error {
	err := k.client.Update(&client.Resource{
		Kind:  "resourcequota",
		Name:  resourceQuota.Name,
		Value: client.NewResourceQuota(resourceQuota),
	}, client.UpdateNamespace(resourceQuota.Namespace))
	if err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("Error updating resource %s: %v", resourceQuota.String(), err)
		}
	}
	return err
}

// deleteResourcequota deletes a resourcequota resource
func (k *kubernetes) deleteResourceQuota(resourceQuota *runtime.ResourceQuota) error {
	err := k.client.Delete(&client.Resource{
		Kind:  "resourcequota",
		Name:  resourceQuota.Name,
		Value: client.NewResourceQuota(resourceQuota),
	}, client.DeleteNamespace(resourceQuota.Namespace))
	if err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("Error deleting resource %s: %v", resourceQuota.String(), err)
		}
	}
	return err
}
