package template

var (
	ProtoType = `syntax = "proto3";
package apis;

option go_package = "{{.Dir}}/proto/apis/{{.Version}};apis";

message {{.Name}}Message {
	string name = 1;
}
`

	ProtoSRV = `syntax = "proto3";

package {{.Name}};

option go_package = "{{.Dir}}/proto/service/{{.Version}}/{{.Name}};{{.Name}}";

service {{title .Name}}Service {
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

	ProtoNew = `syntax = "proto3";

package {{.Name}};

option go_package = "{{.Dir}}/proto/service/{{.Version}}/{{.Name}};{{.Name}}";

service {{title .Name}}Service {
	rpc {{title .Name}}Call({{title .Name}}Request) returns ({{title .Name}}Response) {}
}

message {{title .Name}}Request {
	string name = 1;
}

message {{title .Name}}Response {
	string msg = 1;
}
`
)
