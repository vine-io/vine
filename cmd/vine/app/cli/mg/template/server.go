package template

var (
	SubscriberSRV = `package subscriber

import (
	"context"
	log "github.com/lack-io/vine/lib/logger"

	{{.Name}} "{{.Dir}}/proto/{{.Name}}"
)

type {{title .Alias}} struct{}

func (e *{{title .Alias}}) Handle(ctx context.Context, msg *{{.Name}}.Message) error {
	log.Info("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *{{.Name}}.Message) error {
	log.Info("Function Received message: ", msg.Say)
	return nil
}
`

	HandlerAPI = `package handler

import (
	"context"
	"encoding/json"
	log "github.com/lack-io/vine/lib/logger"

	"{{.Dir}}/client"
	"github.com/lack-io/vine/proto/apis/errors"
	"github.com/lack-io/vine/proto/services/api"
	{{.Name}} "path/to/service/proto/{{.Name}}"
)

type {{title .Alias}} struct{}

func extractValue(pair *api.Pair) string {
	if pair == nil {
		return ""
	}
	if len(pair.Values) == 0 {
		return ""
	}
	return pair.Values[0]
}

// {{title .Alias}}.Call is called by the API as /{{.Name}}/call with post body {"name": "foo"}
func (e *{{title .Alias}}) Call(ctx context.Context, req *api.Request, rsp *api.Response) error {
	log.Info("Received {{title .Alias}}.Call request")

	// extract the client from the context
	{{.Name}}Client, ok := client.{{title .Alias}}FromContext(ctx)
	if !ok {
		return errors.InternalServerError("{{.Alias}}.{{.Name}}.call", "{{.Name}} client not found")
	}

	// make request
	response, err := {{.Name}}Client.Call(ctx, &{{.Name}}.Request{
		Name: extractValue(req.Post["name"]),
	})
	if err != nil {
		return errors.InternalServerError("{{.Alias}}.{{.Name}}.call", err.Error())
	}

	b, _ := json.Marshal(response)

	rsp.StatusCode = 200
	rsp.Body = string(b)

	return nil
}
`

	SingleSRV = `package server

import (
	"context"

	"github.com/lack-io/vine"
	log "github.com/lack-io/vine/lib/logger"

	"{{.Dir}}/pkg/runtime/inject"
	"{{.Dir}}/pkg/service"
	pb "{{.Dir}}/proto/service/{{.Group}}/{{.Version}}"
)

type server struct{
	vine.Service

	H service.{{title .Name}}` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *server) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) error {
	// TODO: Validate
	s.H.Call()
	// FIXME: fix call method
	log.Info("Received {{title .Alias}}.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *server) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
	log.Infof("Received {{title .Name}}.Stream request with count: %d", req.Count)

	// TODO: Validate
	s.H.Stream()
	// FIXME: fix stream method

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&pb.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (s *server) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
	// TODO: Validate
	s.H.PingPong()
	// FIXME: fix stream pingpong

	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&pb.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}

func (s *server) Init(opts ...vine.Option) error {
	var err error
	s.Service.Init(opts...)

	if err = inject.Provide(s.Service, s.Client(), s); err != nil {
		return err
	}

	// TODO: inject more objects

	if err = inject.Populate(); err != nil {
		return err
	}

	if err = s.H.Init(); err != nil {
		return err
	}

	if err = pb.RegisterFooServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New(opts ...vine.Option) *server {
	srv := vine.NewService(opts...)
	return &server{
		Service: srv,
	}
}
`

	ClusterSRV = `package server

import (
	"context"

	"github.com/lack-io/vine"
	log "github.com/lack-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/service"
	"{{.Dir}}/pkg/runtime/inject"
	pb "{{.Dir}}/proto/service/{{.Group}}/{{.Version}}"
)

type server struct{
	vine.Service

	H service.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *server) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) error {
	// TODO: Validate
	s.H.Call()
	// FIXME: fix call method
	log.Info("Received {{title .Name}}.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *server) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
	log.Infof("Received {{title .Name}}.Stream request with count: %d", req.Count)

	// TODO: Validate
	s.H.Stream()
	// FIXME: fix stream method

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&pb.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (s *server) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
	// TODO: Validate
	s.H.PingPong()
	// FIXME: fix stream pingpong

	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&pb.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}

func (s *server) Init(opts ...vine.Option) error {
	var err error
	s.Service.Init(opts...)

	if err = inject.Provide(s.Service, s.Client(), s); err != nil {
		return err
	}

	// TODO: inject more objects

	if err = inject.Populate(); err != nil {
		return err
	}

	if err = s.H.Init(); err != nil {
		return err
	}

	if err = pb.RegisterFooServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New(opts ...vine.Option) *server {
	srv := vine.NewService(opts...)
	return &server{
		Service: srv,
	}
}
`
)
