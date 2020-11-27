package validate

import "github.com/lack-io/vine/cmd/protoc-gen-vine/generator"

func init() {
	generator.RegisterPlugin(new(validate))
}

type validate struct {
	gen *generator.Generator
}

// P forwards to g.gen.P.
func (g *validate) P(args ...interface{}) { g.gen.P(args...) }

func (g *validate) Name() string {
	return "validate"
}

func (g *validate) Init(gen *generator.Generator) {
	g.gen = gen
}

func (g *validate) Generate(file *generator.FileDescriptor) {
	g.P(`// validate Generate`)

}

func (g *validate) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	g.P(`// validate GenerateImports`)
}
