package template

var (
	DockerFNC = `FROM alpine
ADD {{.Alias}}-{{.Type}} /{{.Alias}}-{{.Type}}
ENTRYPOINT [ "/{{.Alias}}-{{.Type}}" ]
`

	DockerSRV = `FROM alpine
ADD {{.Alias}} /{{.Alias}}
ENTRYPOINT [ "/{{.Alias}}" ]
`

	DockerWEB = `FROM alpine
ADD html /html
ADD {{.Alias}}-{{.Type}} /{{.Alias}}-{{.Type}}
WORKDIR /
ENTRYPOINT [ "/{{.Alias}}-{{.Type}}" ]
`
)
