// MIT License
//
// Copyright (c) 2021 Lack
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

package registry

// Service represents a vine service
type Service struct {
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []*Endpoint       `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
	TTL       int64             `json:"ttl,omitempty"`
	Apis      []*OpenAPI        `json:"apis,omitempty"`
}

// Node represents the node the service is on
type Node struct {
	Id       string            `json:"id,omitempty"`
	Address  string            `json:"address,omitempty"`
	Port     int64             `json:"port,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Endpoint is a endpoint provided by a service
type Endpoint struct {
	Name     string            `json:"name,omitempty"`
	Request  *Value            `json:"request,omitempty"`
	Response *Value            `json:"response,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Value is an opaque value for a request or response
type Value struct {
	Name   string   `json:"name,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []*Value `json:"values,omitempty"`
}

// Result is returns by the watcher
type Result struct {
	Action    string   `json:"action,omitempty"`
	Service   *Service `json:"service,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
}

type EventType string

const (
	EventCreate EventType = "Create"
	EventUpdate EventType = "Update"
	EventDelete EventType = "Delete"
)

// Event is registry event
type Event struct {
	// Event Id
	Id string `json:"id,omitempty"`
	// type of event
	Type EventType `json:"type,omitempty"`
	// unix timestamp of event
	Timestamp int64 `json:"timestamp,omitempty"`
	// service entry
	Service *Service `json:"service,omitempty"`
}

type OpenAPI struct {
	Openapi      string                  `json:"openapi,omitempty"`
	Info         *OpenAPIInfo            `json:"info,omitempty"`
	ExternalDocs *OpenAPIExternalDocs    `json:"externalDocs,omitempty"`
	Servers      []*OpenAPIServer        `json:"servers,omitempty"`
	Tags         []*OpenAPITag           `json:"tags,omitempty"`
	Paths        map[string]*OpenAPIPath `json:"paths,omitempty"`
	Components   *OpenAPIComponents      `json:"components,omitempty"`
}

type OpenAPIServer struct {
	Url         string `json:"url,omitempty"`
	Description string `json:"Description,omitempty"`
}

type OpenAPIInfo struct {
	Title          string          `json:"title,omitempty"`
	Description    string          `json:"description,omitempty"`
	TermsOfService string          `json:"termsOfService,omitempty"`
	Contact        *OpenAPIContact `json:"contact,omitempty"`
	License        *OpenAPILicense `json:"license,omitempty"`
	Version        string          `json:"version,omitempty"`
}

type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type OpenAPILicense struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type OpenAPITag struct {
	Name         string               `json:"name,omitempty"`
	Description  string               `json:"description,omitempty"`
	ExternalDocs *OpenAPIExternalDocs `json:"externalDocs,omitempty"`
}

type OpenAPIExternalDocs struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

type OpenAPIPath struct {
	Get    *OpenAPIPathDocs `json:"get,omitempty"`
	Post   *OpenAPIPathDocs `json:"post,omitempty"`
	Put    *OpenAPIPathDocs `json:"put,omitempty"`
	Patch  *OpenAPIPathDocs `json:"patch,omitempty"`
	Delete *OpenAPIPathDocs `json:"delete,omitempty"`
}

type OpenAPIPathDocs struct {
	Tags        []string                 `json:"tags,omitempty"`
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	OperationId string                   `json:"operationId,omitempty"`
	Deprecated  bool                     `json:"deprecated,omitempty"`
	RequestBody *PathRequestBody         `json:"requestBody,omitempty"`
	Parameters  []*PathParameters        `json:"parameters,omitempty"`
	Responses   map[string]*PathResponse `json:"responses,omitempty"`
	Security    []*PathSecurity          `json:"security,omitempty"`
}

type PathSecurity struct {
	Basic   []string `json:"basic,omitempty"`
	ApiKeys []string `json:"apiKeys,omitempty"`
	Bearer  []string `json:"bearer,omitempty"`
}

type PathParameters struct {
	// query, cookie, path
	In              string  `json:"in,omitempty"`
	Name            string  `json:"name,omitempty"`
	Required        bool    `json:"required,omitempty"`
	Description     string  `json:"description,omitempty"`
	AllowReserved   bool    `json:"allowReserved,omitempty"`
	Style           string  `json:"style,omitempty"`
	Explode         bool    `json:"explode,omitempty"`
	AllowEmptyValue bool    `json:"allowEmptyValue,omitempty"`
	Schema          *Schema `json:"schema,omitempty"`
	Example         string  `json:"example,omitempty"`
}

type PathRequestBody struct {
	Description string                  `json:"description,omitempty"`
	Required    bool                    `json:"required,omitempty"`
	Content     *PathRequestBodyContent `json:"content,omitempty"`
}

type PathRequestBodyContent struct {
	ApplicationJson *ApplicationContent `json:"application/json,omitempty"`
	ApplicationXml  *ApplicationContent `json:"application/xml,omitempty"`
}

type ApplicationContent struct {
	Schema *Schema `json:"schema,omitempty"`
}

// PathResponse is swagger path response
type PathResponse struct {
	Description string                  `json:"description,omitempty"`
	Content     *PathRequestBodyContent `json:"content,omitempty"`
}

type OpenAPIComponents struct {
	SecuritySchemes *SecuritySchemes  `json:"securitySchemes,omitempty"`
	Schemas         map[string]*Model `json:"schemas,omitempty"`
}

type SecuritySchemes struct {
	Basic   *BasicSecurity   `json:"basic,omitempty"`
	ApiKeys *APIKeysSecurity `json:"apiKeys,omitempty"`
	Bearer  *BearerSecurity  `json:"bearer,omitempty"`
}

// BasicSecurity is swagger Basic Authorization security (https://swagger.io/docs/specification/authentication/basic-authentication/)
type BasicSecurity struct {
	// http, apiKey, oauth, openIdConnect
	Type   string `json:"type,omitempty"`
	Scheme string `json:"scheme,omitempty"`
}

// APIKeysSecurity is swagger API keys Authorization security (https://swagger.io/docs/specification/authentication/api-keys/)
type APIKeysSecurity struct {
	Type string `json:"type,omitempty"`
	// header
	In   string `json:"in,omitempty"`
	Name string `json:"name,omitempty"`
}

// BearerSecurity is swagger Bearer Authorization security (https://swagger.io/docs/specification/authentication/bearer-authentication/)
type BearerSecurity struct {
	// http
	Type   string `json:"type,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	// JWT
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// Model is swagger data models (https://swagger.io/docs/specification/data-models/)
type Model struct {
	// string, number, integer, boolean, array, object
	Type       string             `json:"type,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`
}

type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Format               string            `json:"format,omitempty"`
	Description          string            `json:"description,omitempty"`
	Example              string            `json:"example,omitempty"`
	Pattern              string            `json:"pattern,omitempty"`
	Nullable             bool              `json:"nullable,omitempty"`
	ReadOnly             bool              `json:"readOnly,omitempty"`
	WriteOnly            bool              `json:"writeOnly,omitempty"`
	Required             bool              `json:"required,omitempty"`
	Ref                  string            `json:"$ref,omitempty"`
	Default              string            `json:"default,omitempty"`
	MinLength            int32             `json:"minLength,omitempty"`
	MaxLength            int32             `json:"maxLength,omitempty"`
	MultipleOf           int32             `json:"multipleOf,omitempty"`
	Minimum              int32             `json:"minimum,omitempty"`
	ExclusiveMinimum     bool              `json:"exclusiveMinimum,omitempty"`
	Maximum              int32             `json:"maximum,omitempty"`
	ExclusiveMaximum     bool              `json:"exclusiveMaximum,omitempty"`
	Enum                 []string          `json:"enum,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	Parameters           []*PathParameters `json:"parameters,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
}
