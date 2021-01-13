package plugin

import (
	"github.com/lack-io/vine/cmd/generator"
)

// validator is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for validator support.
type validator struct {
	gen *generator.Generator
}

func New() *validator {
	return &validator{}
}

// Name returns the name of this plugin, "validator".
func (g *validator) Name() string {
	return "validator"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	pkgImports map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *validator) Init(gen *generator.Generator) {
	g.gen = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *validator) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *validator) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *validator) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *validator) Generate(file *generator.FileDescriptor) {
	for i, msg := range file.Messages() {
		g.generateMessage(file, msg, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *validator) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	// We need to keep track of imported packages to make sure we don't produce
	// a name collision when generating types.
	pkgImports = make(map[generator.GoPackageName]bool)
	for _, name := range imports {
		pkgImports[name] = true
	}
}

func (g *validator) generateMessage(file *generator.FileDescriptor, comment *generator.MessageDescriptor, index int) {
}
