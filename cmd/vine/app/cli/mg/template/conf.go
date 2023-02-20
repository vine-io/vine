package template

var (
	ConfSRV = `server:
	name: {{.Name}}
    address: "127.0.0.1:11401"
`

	ConfGateway = `server:
	name: {{.Name}}
    address: "127.0.0.1:11401"
`
)
