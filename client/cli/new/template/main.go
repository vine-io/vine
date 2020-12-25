package template

var (
	MainFNC = `package main

import (
  	log	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service"
	"{{.Dir}}/handler"
	"{{.Dir}}/subscriber"
)

func main() {
	// New Service
	function := service.NewFunction(
		service.Name("{{.FQDN}}"),
		service.Version("latest"),
	)

	// Initialise function
	function.Init()

	// Register Handler
	function.Handle(new(handler.{{title .Alias}}))

	// Register Struct as Subscriber
	function.Subscribe("{{.FQDN}}", new(subscriber.{{title .Alias}}))

	// Run service
	if err := function.Run(); err != nil {
		log.Fatal(err)
	}
}
`

	MainSRV = `package main

import (
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service"
	"{{.Dir}}/handler"
	"{{.Dir}}/subscriber"

	{{.Alias}} "{{.Dir}}/proto/{{.Alias}}"
)

func main() {
	// New Service
	srv := service.NewService(
		service.Name("{{.FQDN}}"),
		service.Version("latest"),
	)

	// Initialise service
	srv.Init()

	// Register Handler
	{{.Alias}}.Register{{title .Alias}}Handler(service.Server(), new(handler.{{title .Alias}}))

	// Register Struct as Subscriber
	service.RegisterSubscriber("{{.FQDN}}", srv.Server(), new(subscriber.{{title .Alias}}))

	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
`
	MainAPI = `package main

import (
	log "github.com/lack-io/vine/service/logger"

	"github.com/lack-io/vine/service"
	"{{.Dir}}/handler"
	"{{.Dir}}/client"

	{{.Alias}} "{{.Dir}}/proto/{{.Alias}}"
)

func main() {
	// New Service
	srv := service.NewService(
		service.Name("{{.FQDN}}"),
		service.Version("latest"),
	)

	// Initialise service
	srv.Init(
		// create wrap for the {{title .Alias}} service client
		service.WrapHandler(client.{{title .Alias}}Wrapper(srv)),
	)

	// Register Handler
	{{.Alias}}.Register{{title .Alias}}Handler(service.Server(), new(handler.{{title .Alias}}))

	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
`
	MainWEB = `package main

import (
        log "github.com/lack-io/vine/service/logger"
    	"net/http"
        "github.com/lack-io/vine/service/web"
        "{{.Dir}}/handler"
)

func main() {
	// create new web service
        service := web.NewService(
                web.Name("{{.FQDN}}"),
                web.Version("latest"),
        )

	// initialise service
        if err := service.Init(); err != nil {
                log.Fatal(err)
        }

	// register html handler
	service.Handle("/", http.FileServer(http.Dir("html")))

	// register call handler
	service.HandleFunc("/{{.Alias}}/call", handler.{{title .Alias}}Call)

	// run service
        if err := service.Run(); err != nil {
                log.Fatal(err)
        }
}
`
)
