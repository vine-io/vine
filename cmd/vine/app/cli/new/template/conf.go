package template

var (
	ConfSRV = `#[server]
SERVER_NAME="{{.Alias}}"
SERVER_ADDRESS="localhost:11401"
`
)
