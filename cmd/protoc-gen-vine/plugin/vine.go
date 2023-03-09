// Copyright 2020 lack
// MIT License
//
// Copyright (c) 2020 Lack
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

package plugin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vine-io/vine/cmd/generator"
)

// vine is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for vine support.
type vine struct {
	generator.PluginImports
	gen      *generator.Generator
	security *LinkComponents
	schemas  *LinkComponents
	m        sync.Map

	apiPbPkg     generator.Single
	apiPkg       generator.Single
	openApiPkg   generator.Single
	openApiPbPkg generator.Single
	contextPkg   generator.Single
	clientPkg    generator.Single
	serverPkg    generator.Single
}

func New() *vine {
	return &vine{}
}

// Name returns the name of this plugin, "vine".
func (g *vine) Name() string {
	return "vine"
}

// Init initializes the plugin.
func (g *vine) Init(gen *generator.Generator) {
	g.gen = gen
	g.security = NewLinkComponents()
	g.schemas = NewLinkComponents()
	g.PluginImports = generator.NewPluginImports(g.gen)
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

	g.contextPkg = g.NewImport("context", "context")
	g.apiPbPkg = g.NewImport("github.com/vine-io/vine/lib/api", "api")
	g.openApiPkg = g.NewImport("github.com/vine-io/vine/lib/api/handler/openapi", "openapi")
	g.openApiPbPkg = g.NewImport("github.com/vine-io/vine/lib/api/handler/openapi/proto", "openapipb")
	g.apiPkg = g.NewImport("github.com/vine-io/vine/lib/api", "api")
	g.clientPkg = g.NewImport("github.com/vine-io/vine/core/client", "client")
	g.serverPkg = g.NewImport("github.com/vine-io/vine/core/server", "server")

	for i := range file.TagServices() {
		g.generateService(file, file.TagServices()[i], i)
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
	g.P("func New", servName, "Endpoints () []", g.apiPbPkg.Use(), ".Endpoint {")
	g.P("return []", g.apiPbPkg.Use(), ".Endpoint{")
	for _, method := range service.Methods {
		g.generateEndpoint(servName, method)
	}
	g.P("}")
	g.P("}")
	g.P()

	g.P()

	svcTags := g.extractTags(service.Comments)
	if _, ok := svcTags[_openapi]; ok {
		g.P("// Swagger OpenAPI 3.0 for ", servName, " service")
		g.P("func New", servName, "OpenAPI () *", g.openApiPbPkg.Use(), ".OpenAPI {")
		g.P("return &", g.openApiPbPkg.Use(), ".OpenAPI{")
		g.generateOpenAPI(service, svcTags)
		g.P("}")
		g.P("}")
		g.P()
	}

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
	g.P("c ", g.clientPkg.Use(), ".Client")
	g.P("name string")
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("func New", servAlias, " (name string, c ", g.clientPkg.Use(), ".Client) ", servAlias, " {")
	/*
		g.P("if c == nil {")
		g.P("c = ", g.clientPkg.Use(), ".NewClient()")
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
	g.P("func Register", servName, "Handler(s ", g.serverPkg.Use(), ".Server, hdlr ", serverType, ", opts ...", g.serverPkg.Use(), ".HandlerOption) error {")
	g.P("type ", unexport(servName)+"Impl", " interface {")

	// generate interface methods
	for _, method := range service.Methods {
		methName := generator.CamelCase(method.Proto.GetName())
		inType := g.typeName(method.Proto.GetInputType())
		outType := g.typeName(method.Proto.GetOutputType())

		if !method.Proto.GetServerStreaming() && !method.Proto.GetClientStreaming() {
			g.P(methName, "(ctx ", g.contextPkg.Use(), ".Context, in *", inType, ", out *", outType, ") error")
			continue
		}
		g.P(methName, "(ctx ", g.contextPkg.Use(), ".Context, stream server.Stream) error")
	}
	g.P("}")
	g.P("type ", servName, " struct {")
	g.P(unexport(servName) + "Impl")
	g.P("}")
	g.P("h := &", unexport(servName), "Handler{hdlr}")
	g.P(fmt.Sprintf("endpoints := New%sEndpoints()", servName))
	g.P(fmt.Sprintf("for _, ep := range endpoints { opts = append(opts, %s.WithEndpoint(&ep)) }", g.apiPkg.Use()))
	if _, ok := svcTags[_openapi]; ok {
		g.P(fmt.Sprintf(`%s.RegisterOpenAPIDoc(New%sOpenAPI())`, g.openApiPkg.Use(), servName))
		g.P(fmt.Sprintf(`%s.InjectEndpoints(endpoints...)`, g.openApiPkg.Use()))
	}
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
func (g *vine) generateEndpoint(servName string, method *generator.MethodDescriptor) {
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

	g.P("", g.apiPbPkg.Use(), ".Endpoint{")
	desc := extractDesc(method.Comments)
	if len(desc) == 0 {
		desc = []string{servName + "." + method.Proto.GetName()}
	}

	g.P("Name:", fmt.Sprintf(`"%s.%s",`, servName, method.Proto.GetName()))
	g.P("Description:", fmt.Sprintf(`"%s",`, strings.Join(desc, " ")))
	g.P("Path:", fmt.Sprintf(`[]string{"%s"},`, path))
	g.P("Method:", fmt.Sprintf(`[]string{"%s"},`, meth))
	if v, ok := tags[_entity]; ok {
		g.P("Entity:", fmt.Sprintf(`"%s",`, v.Value))
	}
	if v, ok := tags[_security]; ok {
		g.P("Security:", fmt.Sprintf(`"%s",`, v.Value))
	}
	if v, ok := tags[_body]; ok {
		g.P("Body:", fmt.Sprintf(`"%s",`, v.Value))
	} else {
		g.P("Body:", `"*",`)
	}
	if method.Proto.GetServerStreaming() && method.Proto.GetClientStreaming() {
		g.P(fmt.Sprintf("Stream: string(%s.Bidirectional),", g.apiPkg.Use()))
	} else if method.Proto.GetServerStreaming() {
		g.P(fmt.Sprintf("Stream: string(%s.Server),", g.apiPkg.Use()))
	} else if method.Proto.GetClientStreaming() {
		g.P(fmt.Sprintf("Stream: string(%s.Client),", g.apiPkg.Use()))
	}
	g.P(`Handler: "rpc",`)
	defer g.P("},")
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

	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, g.contextPkg.Use(), reqArg, g.clientPkg.Use(), respName)
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
	g.P("stream ", g.clientPkg.Use(), ".Stream")
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
	reqArgs = append(reqArgs, g.contextPkg.Use()+".Context")

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
		g.P("func (h *", unexport(servName), "Handler) ", methName, "(ctx ", g.contextPkg.Use(), ".Context, in *", inType, ", out *", outType, ") error {")
		g.P(fmt.Sprintf(`return h.%s.%s(ctx, in, out)`, serveType, methName))
		g.P("}")
		g.P()
		return hname
	}
	streamType := unexport(servName) + methName + "Stream"
	g.P("func (h *", unexport(servName), "Handler) ", methName, "(ctx ", g.contextPkg.Use(), ".Context, stream server.Stream) error {")
	if !method.Proto.GetClientStreaming() {
		g.P("m := new(", inType, ")")
		g.P("if err := stream.Recv(m); err != nil { return err }")
		g.P(fmt.Sprintf(`return h.%s.%s(ctx, m, &%s{stream})`, serveType, methName, streamType))
	} else {
		g.P(fmt.Sprintf(`return h.%s.%s(ctx, &%s{stream})`, serveType, methName, streamType))
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
	g.P("stream ", g.serverPkg.Use(), ".Stream")
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
