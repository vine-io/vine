package api

import (
	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine"
	ahandler "github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	arpc "github.com/vine-io/vine/lib/api/handler/rpc"
	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/resolver/grpc"
	"github.com/vine-io/vine/lib/api/router"
	regRouter "github.com/vine-io/vine/lib/api/router/registry"
	"github.com/vine-io/vine/util/namespace"
)

func NewRPCGateway(s vine.Service, ns string, wrapper func(*gin.Engine)) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())
	openapi.RegisterOpenAPI(s.Client(), s.Options().Registry, app)

	if wrapper != nil {
		wrapper(app)
	}

	Type := "api"
	HandlerType := "rpc"

	// create the namespace resolver
	nsResolver := namespace.NewResolver(Type, ns)
	// resolver options
	rops := []resolver.Option{
		resolver.WithNamespace(nsResolver.ResolveWithType),
		resolver.WithHandler(HandlerType),
	}

	rr := grpc.NewResolver(rops...)
	rt := regRouter.NewRouter(
		router.WithHandler(arpc.Handler),
		router.WithResolver(rr),
		router.WithRegistry(s.Options().Registry),
	)

	rp := arpc.NewHandler(
		ahandler.WithNamespace(ns),
		ahandler.WithRouter(rt),
		ahandler.WithClient(s.Client()),
	)
	app.Use(rp.Handle)

	return app
}
