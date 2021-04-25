package template

var (
	DockerSRV = `FROM alpine
ADD {{.Name}} /{{.Name}}
ENTRYPOINT [ "/{{.Name}}" ]
`

	DockerWEB = `FROM alpine
ADD html /html
ADD {{.Name}} /{{.Name}}
WORKDIR /
ENTRYPOINT [ "/{{.Name}}" ]
`
)
