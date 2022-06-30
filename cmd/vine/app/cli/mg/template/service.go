package template

var (
	SingleINF = `package service

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/server"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/biz"
	"{{.Dir}}/pkg/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(service))
}

type service struct{
	H biz.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *service) Call(ctx *vine.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *service) Stream(ctx *vine.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *service) PingPong(ctx *vine.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *service) Register(svc server.Server) error {
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	SingleINFWithAPI = `package service

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/server"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/biz"
	"{{.Dir}}/pkg/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(service))
}

type service struct{
	H biz.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *service) Call(ctx *vine.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *service) Stream(ctx *vine.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *service) PingPong(ctx *vine.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *service) Register(svc server.Server) error {
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	ClusterINF = `package service

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/server"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/biz"
	"{{.Dir}}/pkg/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(service))
}

type service struct{
	H biz.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *service) Call(ctx *vine.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *service) Stream(ctx *vine.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *service) PingPong(ctx *vine.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *service) Register(svc server.Server) error {
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`

	ClusterINFWithAPI = `package service

import (
	"github.com/vine-io/apimachinery/inject"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/server"
	verrs "github.com/vine-io/vine/lib/errors"
	log "github.com/vine-io/vine/lib/logger"

	"{{.Dir}}/pkg/{{.Name}}/biz"
	"{{.Dir}}/pkg/version"

	pb "{{.Dir}}/api/services/{{.Group}}/{{.Version}}"
)

func init() {
	inject.ProvidePanic(new(service))
}

type service struct{
	H biz.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *service) Call(ctx *vine.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(version.{{title .Name}}Name, err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *service) Stream(ctx *vine.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *service) PingPong(ctx *vine.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *service) Register(svc server.Server) error {
	return pb.Register{{title .Name}}ServiceHandler(svc, s)
}
`
)
