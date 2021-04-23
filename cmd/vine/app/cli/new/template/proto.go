package template

var (
	ProtoType = `// +dao:output={{.Dir}}/pkg/dao;dao
syntax = "proto3";
package apis;

option go_package = "{{.Dir}}/proto/apis;apis";

message Message {
	string name = 1;
}
`

	ProtoSRV = `syntax = "proto3";

package {{.Name}};

option go_package = "{{.Dir}}/proto/service/{{.Name}};{{.Name}}";

service {{title .Name}} {
	rpc Call(Request) returns (Response) {}
	rpc Stream(StreamingRequest) returns (stream StreamingResponse) {}
	rpc PingPong(stream Ping) returns (stream Pong) {}
}

message Message {
	string say = 1;
}

message Request {
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
)
