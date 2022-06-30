package template

var (
	Module = `module {{.Dir}}

go {{.GoVersion}}

require (
	github.com/vine-io/vine {{.VineVersion}}
	github.com/vine-io/apimachinery {{.VineVersion}}
)

replace google.golang.org/grpc => google.golang.org/grpc v1.47.0
`
)
