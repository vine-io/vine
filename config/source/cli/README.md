# cli Source

The cli source reads config from parsed flags via a cli.Context.

## Format

We expect the use of the `vine/cli` package. Upper case flags will be lower cased. Dashes will be used as delimiters for nesting.

### Example

```go
vine.Flags(
    cli.StringFlag{
        Name: "database-address",
        Value: "127.0.0.1",
        Usage: "the db address",
    },
    cli.IntFlag{
        Name: "database-port",
        Value: 3306,
        Usage: "the db port",
    },
)
```

Becomes

```json
{
    "database": {
        "address": "127.0.0.1",
        "port": 3306
    }
}
```

## New and Load Source

Because a cli.Context is needed to retrieve the flags and their values, it is recommended to build your source from within a cli.Action.

```go

func main() {
    // New Service
    service := vine.NewService(
        vine.Name("example"),
        vine.Flags(
            cli.StringFlag{
                Name: "database-address",
                Value: "127.0.0.1",
                Usage: "the db address",
            },
        ),
    )

    var clisrc source.Source

    service.Init(
        vine.Action(func(c *cli.Context) {
            clisrc = cli.NewSource(
                cli.Context(c),
	    )
            // Alternatively, just setup your config right here
        }),
    )
    
    // ... Load and use that source ...
    conf := config.NewConfig()
    conf.Load(clisrc)
}
```
