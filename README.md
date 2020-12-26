# 安装

```bash
go get github.com/gogo/protobuf
go get github.com/gogo/googleapis
go get github.com/lack-io/vine/cmd/protoc-gen-gogofaster
go get github.com/lack-io/vine/cmd/protoc-gen-vine
go get github.com/lack-io/vine
```

# proto 文件
```protobuf
syntax = "proto3";

package testdata;

import "github.com/gogo/googleapis/google/api/annotations.proto";

service Rpc {
  rpc HelloWorld(HelloWorldRequest) returns (HelloWorldResponse) {
    option (google.api.http) = { post: "/api/{name}"; body: "*"; };
  };
}

message HelloWorldRequest {
  string name = 1;
}

message HelloWorldResponse {
  string reply = 1;
}
```

生成 golang 代码

```bash
protoc -I=. \
  -I=$GOPATH/src \
  -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
  -I=$GOPATH/src/github.com/gogo/googleapis \
  --gogofaster_out=plugins=grpc:. --vine_out=. example/proto/test.proto
```

# 服务端
```go
package main

import (
	"context"
	"fmt"

	vine "github.com/lack-io/vine/service"
	pb "example/proto"
)

type HelloWorld struct {
}

func (t *HelloWorld) HelloWorld(ctx context.Context, req *pb.HelloWorldRequest, rsp *pb.HelloWorldResponse) error {
	fmt.Println("get echo request")
	rsp.Reply = "hello " + req.Name
	return nil
}

func main() {
	service := vine.NewService(
		vine.Name("tt"),
		vine.Address("127.0.0.1:9000"),
	)

	service.Init()

	pb.RegisterRpcHandler(service.Server(), new(HelloWorld))

	service.Run()
}
```
# 客户端
```go
package main

import (
	"context"
	"fmt"

	vine "github.com/lack-io/vine/service"
	pb "example/proto"
)

func main() {
	srv := vine.NewService(vine.Name("tt"))
	service := pb.NewRpcService("tt", srv.Client())

	rsp, err := service.HelloWorld(context.TODO(), &pb.HelloWorldRequest{Name: "world"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("result: %v\n", rsp)
}

```

# 启动 api

```bash
vine api --handler=rpc
```

# 验证

```bash
curl -H 'Content-Type: application/json' -d '{}' http://localhost:8080/api/vine
```