package template

var (
	SingleINF = `package interfaces

import (
	"context"

	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"
	verrs "github.com/vine-io/vine/lib/errors"

	"{{.Dir}}/pkg/runtime"
	"{{.Dir}}/pkg/runtime/inject"
	"{{.Dir}}/pkg/app"
	pb "{{.Dir}}/api/service/{{.Group}}/{{.Version}}"
)

type {{title .Name}}API struct{
	vine.Service

	H app.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *{{title .Name}}API) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(s.Name(), err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *{{title .Name}}API) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *{{title .Name}}API) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *{{title .Name}}API) Init() error {
	var err error

	opts := []vine.Option{
		vine.Name(runtime.{{title .Name}}Name),
		vine.Id(runtime.{{title .Name}}Id),
		vine.Version(runtime.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": runtime.Namespace,
		}),
	}

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

	if err = pb.Register{{title .Name}}ServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New() *{{title .Name}}API {
	srv := vine.NewService()
	return &{{title .Name}}API{
		Service: srv,
	}
}
`

	SingleINFWithAPI = `package interfaces

import (
	"context"

	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"
	verrs "github.com/vine-io/vine/lib/errors"

	"{{.Dir}}/pkg/runtime"
	"{{.Dir}}/pkg/runtime/inject"
	"{{.Dir}}/pkg/app"
	pb "{{.Dir}}/api/service/{{.Group}}/{{.Version}}"
)

type {{title .Name}}API struct{
	vine.Service

	H app.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *{{title .Name}}API) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(s.Name(), err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *{{title .Name}}API) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *{{title .Name}}API) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *{{title .Name}}API) Init() error {
	var err error

	opts := []vine.Option{
		vine.Name(runtime.{{title .Name}}Name),
		vine.Id(runtime.{{title .Name}}Id),
		vine.Version(runtime.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": runtime.Namespace,
		}),
	}

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

	if err = pb.Register{{title .Name}}ServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New() *{{title .Name}}API {
	srv := vine.NewService()
	return &{{title .Name}}API{
		Service: srv,
	}
}
`

	ClusterINF = `package interfaces

import (
	"context"

	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"
	verrs "github.com/vine-io/vine/lib/errors"

	"{{.Dir}}/pkg/runtime"
	"{{.Dir}}/pkg/{{.Name}}/app"
	"{{.Dir}}/pkg/runtime/inject"
	pb "{{.Dir}}/api/service/{{.Group}}/{{.Version}}"
)

type {{title .Name}}API struct{
	vine.Service

	H app.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *{{title .Name}}API) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(s.Name(), err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *{{title .Name}}API) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *{{title .Name}}API) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *{{title .Name}}API) Init() error {
	var err error

	opts := []vine.Option{
		vine.Name(runtime.{{title .Name}}Name),
		vine.Id(runtime.{{title .Name}}Id),
		vine.Version(runtime.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": runtime.Namespace,
		}),
	}

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

	if err = pb.Register{{title .Name}}ServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New() *{{title .Name}}API {
	srv := vine.NewService()
	return &{{title .Name}}API{
		Service: srv,
	}
}
`

	ClusterINFWithAPI = `package interfaces

import (
	"context"

	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"
	verrs "github.com/vine-io/vine/lib/errors"

	"{{.Dir}}/pkg/runtime"
	"{{.Dir}}/pkg/{{.Name}}/app"
	"{{.Dir}}/pkg/runtime/inject"
	pb "{{.Dir}}/api/service/{{.Group}}/{{.Version}}"
)

type {{title .Name}}API struct{
	vine.Service

	H app.{{title .Name}} ` + "`inject:\"\"`" + `
}

// Call is a single request handler called via client.Call or the generated client code
func (s *{{title .Name}}API) Call(ctx context.Context, req *pb.Request, rsp *pb.Response) (err error) {
	if err = req.Validate(); err != nil {
		return verrs.BadRequest(s.Name(), err.Error())
	}
	rsp.Msg, err = s.H.Call(ctx, req.Name)
	return
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (s *{{title .Name}}API) Stream(ctx context.Context, req *pb.StreamingRequest, stream pb.{{title .Name}}Service_StreamStream) error {
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
func (s *{{title .Name}}API) PingPong(ctx context.Context, stream pb.{{title .Name}}Service_PingPongStream) error {
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

func (s *{{title .Name}}API) Init() error {
	var err error

	opts := []vine.Option{
		vine.Name(runtime.{{title .Name}}Name),
		vine.Id(runtime.{{title .Name}}Id),
		vine.Version(runtime.GetVersion()),
		vine.Metadata(map[string]string{
			"namespace": runtime.Namespace,
		}),
	}

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

	if err = pb.Register{{title .Name}}ServiceHandler(s.Service.Server(), s); err != nil {
		return err
	}

	return err
}

func New() *{{title .Name}}API {
	srv := vine.NewService()
	return &{{title .Name}}API{
		Service: srv,
	}
}
`
)
