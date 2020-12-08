package main

import (
	"context"
	"fmt"

	"github.com/lack-io/vine"
	pb "github.com/lack-io/vine/example/proto"
)

type HelloWorld struct {
	pb.RpcHandler
}

func (t HelloWorld) Echo(ctx context.Context, req *pb.HelloWorldRequest, rsp *pb.HelloWorldResponse) error {
	fmt.Println("get echo request")
	rsp.Reply = "hello " + req.Name
	return nil
}

func main() {
	service := vine.NewService(vine.Name("hello-world"))

	service.Init()

	vine.RegisterHandler(service.Server(), new(HelloWorld))

	service.Run()
}
