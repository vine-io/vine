package template

var (
	// Service template is the Vine .mu definition of a service
	Service = `service {{lower .Alias}}
`
)
