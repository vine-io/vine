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

package selector

import (
	"context"

	"github.com/lack-io/vine/service/registry"
)

type Options struct {
	Registry registry.Registry
	Strategy Strategy

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type SelectOptions struct {
	Filters  []Filter
	Strategy Strategy

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Option used to initialise the selector
type Option func(*Options)

// SelectOption used when making a select call
type SelectOption func(*SelectOptions)

// Registry sets the registry used by the selector
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// SetStrategy sets the default strategy for the selector
func SetStrategy(fn Strategy) Option {
	return func(o *Options) {
		o.Strategy = fn
	}
}

// WithFilter adds a filter function to the list of filters
// used during the Select call.
func WithFilter(fn ...Filter) SelectOption {
	return func(o *SelectOptions) {
		o.Filters = append(o.Filters, fn...)
	}
}

// Strategy sets the selector strategy
func WithStrategy(fn Strategy) SelectOption {
	return func(o *SelectOptions) {
		o.Strategy = fn
	}
}
