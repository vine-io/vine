package main

import (
	"context"
	"fmt"

	"github.com/lack-io/vine"
	pb "github.com/lack-io/vine/example/proto"
)

func main() {
	srv := vine.NewService(vine.Name("hello-world"))
	service := pb.NewRpcService("hello-world", srv.Client())

	rsp, err := service.HelloWorld(context.TODO(), &pb.HelloWorldRequest{Name: "world"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("result: %v\n", rsp)
}
