package template

var (
	DockerSRV = `FROM debian:stable-slim
ADD {{.Name}} /{{.Name}}

EXPOSE 11500

ENTRYPOINT [ "/{{.Name}}", "--server-address=0.0.0.0:11500" ]
`

	DockerWEB = `FROM alpine
ADD html /html
ADD {{.Name}} /{{.Name}}
WORKDIR /
ENTRYPOINT [ "/{{.Name}}" ]
`
)
