package api

import (
	"github.com/gin-gonic/gin"
	vclient "github.com/vine-io/vine/core/client"
	ahandler "github.com/vine-io/vine/lib/api/handler"
	"github.com/vine-io/vine/lib/api/handler/openapi"
	arpc "github.com/vine-io/vine/lib/api/handler/rpc"
	"github.com/vine-io/vine/lib/api/resolver"
	"github.com/vine-io/vine/lib/api/resolver/grpc"
	"github.com/vine-io/vine/lib/api/router"
	regRouter "github.com/vine-io/vine/lib/api/router/registry"
	"github.com/vine-io/vine/util/namespace"
)

// PrimpHandler primp *gin.Engine with rpc handler
func PrimpHandler(ns, name string, app *gin.Engine, co vclient.Client, opts ...ahandler.Option) {

	Type := "api"
	HandlerType := "rpc"

	openapi.RegisterOpenAPI(name, co, app)

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
		router.WithRegistry(co.Options().Registry),
	)

	arpcOpts := []ahandler.Option{
		ahandler.WithNamespace(ns),
		ahandler.WithRouter(rt),
		ahandler.WithClient(co),
	}
	arpcOpts = append(arpcOpts, opts...)

	rp := arpc.NewHandler(arpcOpts...)
	app.Use(rp.Handle)
}
