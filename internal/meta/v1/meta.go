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

package v1

import "github.com/lack-io/vine/internal/runtime/schema"

// Object lets you work with object metadata from any of the versioned or
// internal API objects. Attempting to set or retrieve a field on an object that does
// not support that field (Name, UID, Namespace on lists) will be a no-op and return
// a default value.
type Object interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
	GetDesc() string
	SetDesc(desc string)
	GetUID() string
	SetUID(uid string)
	GetCreationTimestamp() int64
	SetCreationTimestamp(timestamp int64)
	GetUpdateTimestamp() int64
	SetUpdateTimestamp(timestamp int64)
	GetDeletionTimestamp() int64
	SetDeletionTimestamp(timestamp int64)
	GetDeletionGrace() bool
	SetDeletionGrace(grace bool)
	GetLabels() Map
	SetLabels(labels Map)
	GetAnnotations() Map
	SetAnnotations(annotations Map)
	GetOwnerReferences() OwnerReferences
	SetOwnerReferences(references OwnerReferences)
}

// ListInterface lets you work with list metadata from any of the versioned or
// internal API Object. Attempting to set or retrieve a field on an object does
// not support that field will be no-op return a default value
type ListInterface interface {
	GetOffset() int32
	SetOffset(offset int32)
	GetLimit() int32
	SetLimit(limit int32)
	GetCount() int64
	SetCount(count int64)
}

func (meta *ListMeta) GetOffset() int32       { return meta.Offset }
func (meta *ListMeta) SetOffset(offset int32) { meta.Offset = offset }
func (meta *ListMeta) GetLimit() int32        { return meta.Limit }
func (meta *ListMeta) SetLimit(limit int32)   { meta.Limit = limit }
func (meta *ListMeta) GetCount() int64        { return meta.Count }
func (meta *ListMeta) SetCount(count int64)   { meta.Count = count }

func (obj *TypeMeta) GetObjectKind() schema.ObjectKind { return obj }

// SetGroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

// GroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

func (obj *ObjectMeta) GetObjectMeta() Object { return obj }

// Namespace implements metav1.Object any object with an ObjectMeta typed field. Allows
// fast, direct access to metadata fields for API objects.
func (meta *ObjectMeta) GetNamespace() string                 { return meta.Namespace }
func (meta *ObjectMeta) SetNamespace(namespace string)        { meta.Namespace = namespace }
func (meta *ObjectMeta) GetName() string                      { return meta.Name }
func (meta *ObjectMeta) SetName(name string)                  { meta.Name = name }
func (meta *ObjectMeta) GetDesc() string                      { return meta.Desc }
func (meta *ObjectMeta) SetDesc(desc string)                  { meta.Desc = desc }
func (meta *ObjectMeta) GetUID() string                       { return meta.UID }
func (meta *ObjectMeta) SetUID(uid string)                    { meta.UID = uid }
func (meta *ObjectMeta) GetCreationTimestamp() int64          { return meta.CreationTimestamp }
func (meta *ObjectMeta) SetCreationTimestamp(timestamp int64) { meta.CreationTimestamp = timestamp }
func (meta *ObjectMeta) GetUpdateTimestamp() int64            { return meta.UpdateTimestamp }
func (meta *ObjectMeta) SetUpdateTimestamp(timestamp int64)   { meta.UpdateTimestamp = timestamp }
func (meta *ObjectMeta) GetDeletionTimestamp() int64          { return meta.DeletionTimestamp }
func (meta *ObjectMeta) SetDeletionTimestamp(timestamp int64) { meta.DeletionTimestamp = timestamp }
func (meta *ObjectMeta) GetDeletionGrace() bool               { return meta.DeletionGrace }
func (meta *ObjectMeta) SetDeletionGrace(grace bool)          { meta.DeletionGrace = grace }
func (meta *ObjectMeta) GetLabels() Map                       { return meta.Labels }
func (meta *ObjectMeta) SetLabels(labels Map)                 { meta.Labels = labels }
func (meta *ObjectMeta) GetAnnotations() Map                  { return meta.Annotations }
func (meta *ObjectMeta) SetAnnotations(annotations Map)       { meta.Annotations = annotations }
func (meta *ObjectMeta) GetOwnerReferences() OwnerReferences  { return meta.OwnerReferences }
func (meta *ObjectMeta) SetOwnerReferences(references OwnerReferences) {
	meta.OwnerReferences = references
}
