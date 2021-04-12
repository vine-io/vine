[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lack/vine?tab=doc) [![Travis CI](https://api.travis-ci.org/lack-io/vine.svg?branch=master)](https://travis-ci.org/lack-io/vine) [![Go Report Card](https://goreportcard.com/badge/lack-io/vine)](https://goreportcard.com/report/github.com/lack-io/vine)

Vine (*/vaɪn/*) 一套简单、高效、可插拔的分布式 RPC 框架。

## 简介

**Vine** 基于 [go-micro](https://github.com/asim/go-micro) 框架，因此继承了它的优点。并在此基础上增加性能和可用性。

## 主要功能

**Vine** 抽象化分布式系统的内部细节. 以下是主要功能.

- **身份验证 (Authentication)** - 身份验证和授权作为一等公民而存在。身份验证和授权通过为每个服务提供一个身份和证书来确保内部服务安全。其中海包括基于规则的访问控制。

- **动态配置 (Dynamic Config)** - 从任意地方加载和热加载动态配置。配置接口提供了一种方法，可以从任何源（如环境变量，文件，etcd）加载应用级别配置。同时支持合并源、甚至定义回退。

- **数据存储 (Data Storage)** - 一个简单的数据存储接口，用于查询，创建，删除数据记录。内置 memory，file，postgresSQL 等

- **服务发现 (Service disconvery)** - 自动服务注册和名称解析。服务发现是微服务开发的核心功能。当服务A与服务B通讯时，需要知道服务B的IP地址等信息。**Vine** 内置(mdns)作为服务发现组件，它是零配置系统。

- **负载均衡 (Load Balancing)** - 基于服务发现的客户端负载均衡。当内部存在多个服务实例的地址时，我们需要一种方式确定路由到哪个节点，默认使用随机指定其中一个地址。同时在请求错误时重试不同的节点。

- **消息编码 (Message Encoding)** - 基于`Content-Type`的动态消息编码。Codec 为客户端和服务端提供 Go 类型的编码和解码，支持各种不同的类型。默认使用 protobuf 和 json。

- **gRPC 传输 (gRPC transport)** - 基于 gRPC 的请求响应，同时支持双向流。**Vine** 为同步通讯提供一个抽象，使发向服务的请求被自动解析、负载均衡、拨号和流化。

- **异步消息 (Async Messaging)** - 内置订阅发布模型，作为异步通讯和事件驱动架构。事件通知属于微服务开发中的核心模式。默认的消息系统为 HTTP 事件代理。

- **同步 (Synchronization)** - 分布式系统通常以最终一致性的方式构建。内置的 **Sync** 接口实现分布式锁和领导选举。

- **可插拔接口 (Pluggable Interfaces)** - 得益于 Go 语言的抽象特性。 **Vine** 为每个模块提供抽象接口，正因为如此，这些接口都是可插拔的。可以在 [github.com/lack-io/plugins](https://github.com/lack-io/plugins) 查询你需要的插件。

## 许可

Vine 遵守 Apache 2.0 开源许可.

## 起步

### 安装

```bash
go get github.com/gogo/protobuf
go get github.com/lack-io/vine/cmd/protoc-gen-gogo
go get github.com/lack-io/vine/cmd/protoc-gen-vine
go get github.com/lack-io/cmd/vine
```

### proto 文件
```protobuf
syntax = "proto3";

package testdata;

// +gen:openapi
service Rpc {
  // +gen:post=/api/{name}
  // +gen:body=*
  rpc HelloWorld(HelloWorldRequest) returns (HelloWorldResponse) {};
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
  --gogo_out=:. --vine_out=. example/proto/test.proto
```

### 服务端
```go
package main

import (
	"context"
	"fmt"

	"github.com/lack-io/vine"
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

### 客户端
```go
package main

import (
	"context"
	"fmt"

	"github.com/lack-io/vine"
	pb "example/proto"
)

func main() {
	svc := vine.NewService(vine.Name("tt"))
	service := pb.NewRpcService("tt", svc.Client())

	rsp, err := service.HelloWorld(context.TODO(), &pb.HelloWorldRequest{Name: "world"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("result: %v\n", rsp)
}

```

### 启动 api

```bash
vine api --handler=rpc
```

### 验证

```bash
curl -H 'Content-Type: application/json' -d '{}' http://localhost:8080/api/vine
```

## 接下来的事
- [x] 修改 github.com/lack-io/vine/service 结构
- [x] 添加内置定时任务功能
- [x] 使用高性能的 web 库替换 mux
- [ ] 修改 Auth 组件
- [x] 减少对应依赖，如 ACEM
- [x] 添加数据库支持
- [ ] 修改 vine new 功能，添加 gateway 模板
- [x] 组件下的 struct 整合进 proto 中，同 registry 的实现
- [ ] 添加 protoc-gen-cli 生成服务的客户端命令行
- [ ] 提供接口，将 openapi 输出 pdf，Markdown 格式
- [ ] 完善 Vine 文档
- [ ] 服务热重启
- [ ] 服务自动升级
- [ ] 部署工具
- [ ] 内置依赖注入
- [ ] 内置 pprof