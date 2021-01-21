// Copyright 2020 lack
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

package plugin

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/lack-io/vine/cmd/generator"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath  = "context"
	apiPkgPath      = "github.com/lack-io/vine/service/api"
	clientPkgPath   = "github.com/lack-io/vine/service/client"
	serverPkgPath   = "github.com/lack-io/vine/service/server"
	registryPkgPath = "github.com/lack-io/vine/proto/registry"
)

// vine is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for vine support.
type vine struct {
	gen        *generator.Generator
	security   map[string]*Component
	schemas    map[string]*Component
	extSchemas map[string]*Component
	errors     map[string]*Component
	m          sync.Map
}

func New() *vine {
	return &vine{}
}

// Name returns the name of this plugin, "vine".
func (g *vine) Name() string {
	return "vine"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	apiPkg      string
	contextPkg  string
	clientPkg   string
	serverPkg   string
	registryPkg string
	pkgImports  map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *vine) Init(gen *generator.Generator) {
	g.gen = gen
	g.security = map[string]*Component{}
	g.schemas = map[string]*Component{}
	g.extSchemas = map[string]*Component{}
	g.errors = map[string]*Component{}
	contextPkg = generator.RegisterUniquePackageName("context", nil)
	apiPkg = generator.RegisterUniquePackageName("api", nil)
	clientPkg = generator.RegisterUniquePackageName("client", nil)
	serverPkg = generator.RegisterUniquePackageName("server", nil)
	registryPkg = generator.RegisterUniquePackageName("registry", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *vine) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *vine) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *vine) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *vine) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", apiPkg, ".Endpoint")
	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", clientPkg, ".Option")
	g.P("var _ ", serverPkg, ".Option")
	g.P("var _ ", registryPkg, ".OpenAPI")
	g.P()

	for i, service := range file.TagServices() {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *vine) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P(contextPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, contextPkgPath)))
	g.P(apiPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, apiPkgPath)))
	g.P(clientPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, clientPkgPath)))
	g.P(serverPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, serverPkgPath)))
	g.P(registryPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, registryPkgPath)))
	g.P(")")
	g.P()

	// We need to keep track of imported packages to make sure we don't produce
	// a name collision when generating types.
	pkgImports = make(map[generator.GoPackageName]bool)
	for _, name := range imports {
		pkgImports[name] = true
	}
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
	// TODO: do we need any in vine?
}

func unexport(s string) string {
	if len(s) == 0 {
		return ""
	}
	name := strings.ToLower(s[:1]) + s[1:]
	if pkgImports[generator.GoPackageName(name)] {
		return name + "_"
	}
	return name
}

// generateService generates all the code for the named service.
func (g *vine) generateService(file *generator.FileDescriptor, service *generator.ServiceDescriptor, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.Proto.GetName()
	serviceName := strings.ToLower(service.Proto.GetName())
	if pkg := file.GetPackage(); pkg != "" {
		serviceName = pkg
	}
	servName := generator.CamelCase(origServName)
	servAlias := servName + "Service"

	// strip suffix
	if strings.HasSuffix(servAlias, "ServiceService") {
		servAlias = strings.TrimSuffix(servAlias, "Service")
	}

	g.P()
	g.P("// API Endpoints for ", servName, " service")
	g.P("func New", servName, "Endpoints () []*", apiPkg, ".Endpoint {")
	g.P("return []*", apiPkg, ".Endpoint{")
	for _, method := range service.Methods {
		g.generateEndpoint(servName, method, false)
	}
	g.P("}")
	g.P("}")
	g.P()

	g.P()
	g.P("// Swagger OpenAPI 3.0 for ", servName, " service")
	g.P("func New", servName, "OpenAPI () *", registryPkg, ".OpenAPI {")
	g.P("return &", registryPkg, ".OpenAPI{")
	g.generateOpenAPI(service)
	g.P("}")
	g.P("}")
	g.P()

	g.P()
	g.P("// Client API for ", servName, " service")
	// Client interface.
	g.gen.PrintComments(fmt.Sprintf("6,%d", index))
	g.P("type ", servAlias, " interface {")
	for i, method := range service.Methods {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateClientSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unexport(servAlias), " struct {")
	g.P("c ", clientPkg, ".Client")
	g.P("name string")
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func New", servAlias, " (name string, c ", clientPkg, ".Client) ", servAlias, " {")
	/*
		g.P("if c == nil {")
		g.P("c = ", clientPkg, ".NewClient()")
		g.P("}")
		g.P("if len(name) == 0 {")
		g.P(`name = "`, serviceName, `"`)
		g.P("}")
	*/
	g.P("return &", unexport(servAlias), "{")
	g.P("c: c,")
	g.P("name: name,")
	g.P("}")
	g.P("}")
	g.P()
	var methodIndex, streamIndex int
	serviceDescVar := "_" + servName + "_serviceDesc"
	// Client method implementations.
	for _, method := range service.Methods {
		var descExpr string
		if !method.Proto.GetServerStreaming() {
			// Unary RPC method
			descExpr = fmt.Sprintf("&%s.Methods[%d]", serviceDescVar, methodIndex)
			methodIndex++
		} else {
			// Streaming RPC method
			descExpr = fmt.Sprintf("&%s.Streams[%d]", serviceDescVar, streamIndex)
			streamIndex++
		}
		g.generateClientMethod(serviceName, servName, serviceDescVar, method, descExpr)
	}

	g.P("// Server API for ", servName, " service")
	// Server interface.
	serverType := servName + "Handler"
	g.gen.PrintComments(fmt.Sprintf("6,%d", index))
	g.P("type ", serverType, " interface {")
	for i, method := range service.Methods {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateServerSignature(servName, method))
	}
	g.P("}")
	g.P()

	// Server registration.
	g.P("func Register", servName, "Handler(s ", serverPkg, ".Server, hdlr ", serverType, ", opts ...", serverPkg, ".HandlerOption) error {")
	g.P("type ", unexport(servName)+"Impl", " interface {")

	// generate interface methods
	for _, method := range service.Methods {
		methName := generator.CamelCase(method.Proto.GetName())
		inType := g.typeName(method.Proto.GetInputType())
		outType := g.typeName(method.Proto.GetOutputType())

		if !method.Proto.GetServerStreaming() && !method.Proto.GetClientStreaming() {
			g.P(methName, "(ctx ", contextPkg, ".Context, in *", inType, ", out *", outType, ") error")
			continue
		}
		g.P(methName, "(ctx ", contextPkg, ".Context, stream server.Stream) error")
	}
	g.P("}")
	g.P("type ", servName, " struct {")
	g.P(unexport(servName) + "Impl")
	g.P("}")
	g.P("h := &", unexport(servName), "Handler{hdlr}")
	for _, method := range service.Methods {
		g.generateEndpoint(servName, method, true)
	}
	g.P("opts = append(opts, server.OpenAPIHandler(New", servName, "OpenAPI()))")
	g.P("return s.Handle(s.NewHandler(&", servName, "{h}, opts...))")
	g.P("}")
	g.P()

	g.P("type ", unexport(servName), "Handler struct {")
	g.P(serverType)
	g.P("}")

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Methods {
		hname := g.generateServerMethod(servName, method)
		handlerNames = append(handlerNames, hname)
	}
}

// generateEndpoint creates the api endpoint
func (g *vine) generateEndpoint(servName string, method *generator.MethodDescriptor, with bool) {
	tags := g.extractTags(method.Comments)
	if len(tags) == 0 {
		return
	}
	var meth string
	var path string

	if v, ok := tags[_get]; ok {
		meth = "GET"
		path = v.Value
	} else if v, ok = tags[_patch]; ok {
		meth = "PATCH"
		path = v.Value
	} else if v, ok = tags[_put]; ok {
		meth = "PUT"
		path = v.Value
	} else if v, ok = tags[_post]; ok {
		meth = "POST"
		path = v.Value
	} else if v, ok = tags[_delete]; ok {
		meth = "DELETE"
		path = v.Value
	} else {
		return
	}
	if with {
		g.P("opts = append(opts, ", apiPkg, ".WithEndpoint(&", apiPkg, ".Endpoint{")
		defer g.P("}))")
	} else {
		g.P("&", apiPkg, ".Endpoint{")
		defer g.P("},")
	}
	g.P("Name:", fmt.Sprintf(`"%s.%s",`, servName, method.Proto.GetName()))
	g.P("Path:", fmt.Sprintf(`[]string{"%s"},`, path))
	g.P("Method:", fmt.Sprintf(`[]string{"%s"},`, meth))
	if v, ok := tags[_body]; ok {
		g.P("Body:", fmt.Sprintf(`"%s",`, v.Value))
	}
	if method.Proto.GetServerStreaming() || method.Proto.GetClientStreaming() {
		g.P("Stream: true,")
	}
	g.P(`Handler: "rpc",`)
}

// generateClientSignature returns the client-side signature for a method.
func (g *vine) generateClientSignature(servName string, method *generator.MethodDescriptor) string {
	origMethName := method.Proto.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}
	reqArg := ", in *" + g.typeName(method.Proto.GetInputType())
	if method.Proto.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + g.typeName(method.Proto.GetOutputType())
	if method.Proto.GetServerStreaming() || method.Proto.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(origMethName) + "Service"
	}

	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, contextPkg, reqArg, clientPkg, respName)
}

func (g *vine) generateClientMethod(reqServ, servName, serviceDescVar string, method *generator.MethodDescriptor, descExpr string) {
	reqMethod := fmt.Sprintf("%s.%s", servName, method.Proto.GetName())
	methName := generator.CamelCase(method.Proto.GetName())
	inType := g.typeName(method.Proto.GetInputType())
	outType := g.typeName(method.Proto.GetOutputType())

	servAlias := servName + "Service"

	// strip suffix
	if strings.HasSuffix(servAlias, "ServiceService") {
		servAlias = strings.TrimSuffix(servAlias, "Service")
	}

	g.P("func (c *", unexport(servAlias), ") ", g.generateClientSignature(servName, method), "{")
	if !method.Proto.GetServerStreaming() && !method.Proto.GetClientStreaming() {
		g.P(`req := c.c.NewRequest(c.name, "`, reqMethod, `", in)`)
		g.P("out := new(", outType, ")")
		// TODO: Pass descExpr to Invoke.
		g.P("err := ", `c.c.Call(ctx, req, out, opts...)`)
		g.P("if err != nil { return nil, err }")
		g.P("return out, nil")
		g.P("}")
		g.P()
		return
	}
	streamType := unexport(servAlias) + methName
	g.P(`req := c.c.NewRequest(c.name, "`, reqMethod, `", &`, inType, `{})`)
	g.P("stream, err := c.c.Stream(ctx, req, opts...)")
	g.P("if err != nil { return nil, err }")

	if !method.Proto.GetClientStreaming() {
		g.P("if err := stream.Send(in); err != nil { return nil, err }")
	}

	g.P("return &", streamType, "{stream}, nil")
	g.P("}")
	g.P()

	genSend := method.Proto.GetClientStreaming()
	genRecv := method.Proto.GetServerStreaming()

	// Stream auxiliary types and methods.
	g.P("type ", servName, "_", methName, "Service interface {")
	g.P("Context() context.Context")
	g.P("SendMsg(interface{}) error")
	g.P("RecvMsg(interface{}) error")

	if genSend && !genRecv {
		// client streaming, the server will send a response upon close
		g.P("CloseAndRecv() (*", outType, ", error)")
	} else {
		g.P("Close() error")
	}

	if genSend {
		g.P("Send(*", inType, ") error")
	}
	if genRecv {
		g.P("Recv() (*", outType, ", error)")
	}
	g.P("}")
	g.P()

	g.P("type ", streamType, " struct {")
	g.P("stream ", clientPkg, ".Stream")
	g.P("}")
	g.P()

	if genSend && !genRecv {
		// client streaming, the server will send a response upon close
		g.P("func (x *", streamType, ") CloseAndRecv() (*", outType, ", error) {")
		g.P("if err := x.stream.Close(); err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("r := new(", outType, ")")
		g.P("err := x.RecvMsg(r)")
		g.P("return r, err")
		g.P("}")
		g.P()
	} else {
		g.P("func (x *", streamType, ") Close() error {")
		g.P("return x.stream.Close()")
		g.P("}")
		g.P()
	}

	g.P("func (x *", streamType, ") Context() context.Context {")
	g.P("return x.stream.Context()")
	g.P("}")
	g.P()

	g.P("func (x *", streamType, ") SendMsg(m interface{}) error {")
	g.P("return x.stream.Send(m)")
	g.P("}")
	g.P()

	g.P("func (x *", streamType, ") RecvMsg(m interface{}) error {")
	g.P("return x.stream.Recv(m)")
	g.P("}")
	g.P()

	if genSend {
		g.P("func (x *", streamType, ") Send(m *", inType, ") error {")
		g.P("return x.stream.Send(m)")
		g.P("}")
		g.P()

	}

	if genRecv {
		g.P("func (x *", streamType, ") Recv() (*", outType, ", error) {")
		g.P("m := new(", outType, ")")
		g.P("err := x.stream.Recv(m)")
		g.P("if err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("return m, nil")
		g.P("}")
		g.P()
	}
}

// generateServerSignature returns the server-side signature for a method.
func (g *vine) generateServerSignature(servName string, method *generator.MethodDescriptor) string {
	origMethName := method.Proto.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}

	var reqArgs []string
	ret := "error"
	reqArgs = append(reqArgs, contextPkg+".Context")

	if !method.Proto.GetClientStreaming() {
		reqArgs = append(reqArgs, "*"+g.typeName(method.Proto.GetInputType()))
	}
	if method.Proto.GetServerStreaming() || method.Proto.GetClientStreaming() {
		reqArgs = append(reqArgs, servName+"_"+generator.CamelCase(origMethName)+"Stream")
	}
	if !method.Proto.GetClientStreaming() && !method.Proto.GetServerStreaming() {
		reqArgs = append(reqArgs, "*"+g.typeName(method.Proto.GetOutputType()))
	}
	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

func (g *vine) generateServerMethod(servName string, method *generator.MethodDescriptor) string {
	methName := generator.CamelCase(method.Proto.GetName())
	hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
	serveType := servName + "Handler"
	inType := g.typeName(method.Proto.GetInputType())
	outType := g.typeName(method.Proto.GetOutputType())

	if !method.Proto.GetServerStreaming() && !method.Proto.GetClientStreaming() {
		g.P("func (h *", unexport(servName), "Handler) ", methName, "(ctx ", contextPkg, ".Context, in *", inType, ", out *", outType, ") error {")
		g.P("return h.", serveType, ".", methName, "(ctx, in, out)")
		g.P("}")
		g.P()
		return hname
	}
	streamType := unexport(servName) + methName + "Stream"
	g.P("func (h *", unexport(servName), "Handler) ", methName, "(ctx ", contextPkg, ".Context, stream server.Stream) error {")
	if !method.Proto.GetClientStreaming() {
		g.P("m := new(", inType, ")")
		g.P("if err := stream.Recv(m); err != nil { return err }")
		g.P("return h.", serveType, ".", methName, "(ctx, m, &", streamType, "{stream})")
	} else {
		g.P("return h.", serveType, ".", methName, "(ctx, &", streamType, "{stream})")
	}
	g.P("}")
	g.P()

	genSend := method.Proto.GetServerStreaming()
	genRecv := method.Proto.GetClientStreaming()

	// Stream auxiliary types and methods.
	g.P("type ", servName, "_", methName, "Stream interface {")
	g.P("Context() context.Context")
	g.P("SendMsg(interface{}) error")
	g.P("RecvMsg(interface{}) error")

	if !genSend {
		// client streaming, the server will send a response upon close
		g.P("SendAndClose(*", outType, ")  error")
	} else {
		g.P("Close() error")
	}

	if genSend {
		g.P("Send(*", outType, ") error")
	}

	if genRecv {
		g.P("Recv() (*", inType, ", error)")
	}

	g.P("}")
	g.P()

	g.P("type ", streamType, " struct {")
	g.P("stream ", serverPkg, ".Stream")
	g.P("}")
	g.P()

	if !genSend {
		// client streaming, the server will send a response upon close
		g.P("func (x *", streamType, ") SendAndClose(in *", outType, ") error {")
		g.P("if err := x.SendMsg(in); err != nil {")
		g.P("return err")
		g.P("}")
		g.P("return x.stream.Close()")
		g.P("}")
		g.P()
	} else {
		// other types of rpc don't send a response when the stream closes
		g.P("func (x *", streamType, ") Close() error {")
		g.P("return x.stream.Close()")
		g.P("}")
		g.P()
	}

	g.P("func (x *", streamType, ") Context() context.Context {")
	g.P("return x.stream.Context()")
	g.P("}")
	g.P()

	g.P("func (x *", streamType, ") SendMsg(m interface{}) error {")
	g.P("return x.stream.Send(m)")
	g.P("}")
	g.P()

	g.P("func (x *", streamType, ") RecvMsg(m interface{}) error {")
	g.P("return x.stream.Recv(m)")
	g.P("}")
	g.P()

	if genSend {
		g.P("func (x *", streamType, ") Send(m *", outType, ") error {")
		g.P("return x.stream.Send(m)")
		g.P("}")
		g.P()
	}

	if genRecv {
		g.P("func (x *", streamType, ") Recv() (*", inType, ", error) {")
		g.P("m := new(", inType, ")")
		g.P("if err := x.stream.Recv(m); err != nil { return nil, err }")
		g.P("return m, nil")
		g.P("}")
		g.P()
	}

	return hname
}
