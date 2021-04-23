package template

var (
	ConfSRV = `#[server]
SERVER_NAME="{{.Name}}"
SERVER_ADDRESS="localhost:11401"
`

	ConfGateway = `#[server]
SERVER_NAME="{{.Name}}"
SERVER_ADDRESS="localhost:80"
`
)
