package template

var (
	SingleServer = `package server

import (
	"context"

	"github.com/vine-io/apimachinery/inject"
	vserver "github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/service"
	"{{.Dir}}/pkg/internal/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(server))
}

type server struct{
	H service.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *server) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
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

func (s *server) Register(svc vserver.Server) error {
	if err := openapi.RegisterOpenAPIHandler(svc); err != nil {
		return err
	}
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	SingleServerWithAPI = `package server

import (
	"context"

	"github.com/vine-io/apimachinery/inject"
	vserver "github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/service"
	"{{.Dir}}/pkg/internal/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(server))
}

type server struct{
	H service.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *server) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
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

func (s *server) Register(svc vserver.Server) error {
	if err := openapi.RegisterOpenAPIHandler(svc); err != nil {
		return err
	}
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	ClusterServer = `package server

import (
	"context"

	"github.com/vine-io/apimachinery/inject"
	vserver "github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/service"
	"{{.Dir}}/pkg/internal/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(server))
}

type server struct{
	H service.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *server) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
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

func (s *server) Register(svc vserver.Server) error {
	if err := openapi.RegisterOpenAPIHandler(svc); err != nil {
		return err
	}
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	ClusterServerWithAPI = `package server

import (
	"context"

	"github.com/vine-io/apimachinery/inject"
	vserver "github.com/vine-io/vine/core/server"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/biz"
	"{{.Dir}}/pkg/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(server))
}

type server struct{
	H service.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *service) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
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

func (s *server) Register(svc vserver.Server) error {
	if err := openapi.RegisterOpenAPIHandler(svc); err != nil {
		return err
	}
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`
)
