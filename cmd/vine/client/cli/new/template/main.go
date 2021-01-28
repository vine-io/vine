package template

var (
	MainFNC = `package main

import (
  	log	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine"

	"{{.Dir}}/handler"
	"{{.Dir}}/subscriber"
)

func main() {
	// New Service
	function := vine.NewFunction(
		vine.Name("{{.FQDN}}"),
		vine.Version("latest"),
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
	"github.com/lack-io/vine"

	"{{.Dir}}/handler"
	"{{.Dir}}/subscriber"

	{{.Alias}} "{{.Dir}}/proto/{{.Alias}}"
)

func main() {
	// New Service
	service := vine.NewService(
		vine.Name("{{.FQDN}}"),
		vine.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	{{.Alias}}.Register{{title .Alias}}Handler(service.Server(), new(handler.{{title .Alias}}))

	// Register Struct as Subscriber
	vine.RegisterSubscriber("{{.FQDN}}", service.Server(), new(subscriber.{{title .Alias}}))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
`
	MainAPI = `package main

import (
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine"

	"{{.Dir}}/handler"
	"{{.Dir}}/client"
	{{.Alias}} "{{.Dir}}/proto/{{.Alias}}"
)

func main() {
	// New Service
	service := vine.NewService(
		vine.Name("{{.FQDN}}"),
		vine.Version("latest"),
	)

	// Initialise service
	service.Init(
		// create wrap for the {{title .Alias}} service client
		vine.WrapHandler(client.{{title .Alias}}Wrapper(svc)),
	)

	// Register Handler
	{{.Alias}}.Register{{title .Alias}}Handler(service.Server(), new(handler.{{title .Alias}}))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
`
	MainWEB = `package main

import (
    	"net/http"

        log "github.com/lack-io/vine/service/logger"
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
