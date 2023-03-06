package template

var (
	ProtoType = `syntax = "proto3";

package {{.Group}}{{.Version}};

option go_package = "{{.Dir}}/api/types/{{.Group}}/{{.Version}};{{.Group}}{{.Version}}";
option java_package = "io.vine.types.{{.Group}}.{{.Version}}";
option java_multiple_files = true;

import "github.com/vine-io/apimachinery/apis/meta/v1/generated.proto";

// +gen:object
message {{title .Name}} {
	// +gen:inline
	v1.TypeMeta typeMeta = 1;
	// +gen:inline
	v1.ObjectMeta metadata = 2;
	
	// +gen:required
	{{title .Name}}Spec spec = 3;
}

message {{title .Name}}Spec {
	
}
`

	ProtoSRV = `syntax = "proto3";

package {{.Group}}{{.Version}};

option go_package = "{{.Dir}}/api/services/{{.Group}}/{{.Version}};{{.Group}}{{.Version}}";
option java_package = "io.vine.services.{{.Group}}.{{.Version}}";
option java_multiple_files = true;

// +gen:openapi
service {{title .Name}}Service {
	// +gen:entity={{.Name}}
	// +gen:post=/api/{{.Group}}/{{.Version}}/{{.Name}}/Call
	rpc Call(Request) returns (Response);
	rpc Stream(StreamingRequest) returns (stream StreamingResponse);
	rpc PingPong(stream Ping) returns (stream Pong);
}

message Request {
    // +gen:required
	string name = 1;
}

message Response {
	string msg = 1;
}

message StreamingRequest {
	int64 count = 1;
}

message StreamingResponse {
	int64 count = 1;
}

message Ping {
	int64 stroke = 1;
}

message Pong {
	int64 stroke = 1;
}
`

	ProtoNew = `syntax = "proto3";

package {{.Group}}{{.Version}};

option go_package = "{{.Dir}}/api/services/{{.Group}}/{{.Version}};{{.Group}}{{.Version}}";
option java_package = "io.vine.services.{{.Group}}.{{.Version}}";
option java_multiple_files = true;

// +gen:openapi
service {{title .Name}}Service {
	// +gen:post=/api/{{.Group}}/{{.Version}}/{{.Name}}/Call
	rpc {{title .Name}}Call({{title .Name}}Request) returns ({{title .Name}}Response) {}
}

message {{title .Name}}Request {
    // +gen:required
	string name = 1;
}

message {{title .Name}}Response {
	string msg = 1;
}
`

	Register = `package {{.Group}}{{.Version}}

import (
	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/schema"
	"github.com/vine-io/apimachinery/storage"
)

// GroupName is the group name for this API
const GroupName = {{if eq .Group "core"}}""{{else}}"{{.Group}}"{{end}}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "{{.Version}}"}

var (
	SchemaBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemaBuilder.AddToScheme
	sets          = make([]runtime.Object, 0)
)

var (
	FactoryBuilder = storage.NewFactoryBuilder(addKnownFactory)
	AddToBuilder   = FactoryBuilder.AddToFactory
	storageSet     = make([]storage.Storage, 0)
)

func addKnownFactory(f storage.Factory) error {
	return f.AddKnownStorages(SchemeGroupVersion, storageSet...)
}

func addKnownTypes(scheme runtime.Scheme) error {
	return scheme.AddKnownTypes(SchemeGroupVersion, sets...)
}`
)
