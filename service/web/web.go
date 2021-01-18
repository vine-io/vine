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

// Package web provides web based vine services
package web

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Service is a web service with service discovery built in
type Service interface {
	Client() *http.Client
	Init(opts ...Option) error
	Options() Options
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Run() error
}

//Option for web
type Option func(o *Options)

//Web basic Defaults
var (
	// For serving
	DefaultName    = "go-web"
	DefaultVersion = "latest"
	DefaultId      = uuid.New().String()
	DefaultAddress = ":0"

	// for registration
	DefaultRegisterTTL      = time.Second * 90
	DefaultRegisterInterval = time.Second * 30

	// static directory
	DefaultStaticDir     = "html"
	DefaultRegisterCheck = func(context.Context) error { return nil }
)

// NewService returns a new web.Service
func NewService(opts ...Option) Service {
	return newService(opts...)
}
