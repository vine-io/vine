package template

var (
	SystemedSRV = `[Unit]
Description={{.Alias}}
After=network.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/vine/{{.Name}}
ExecStart=/opt/vine/{{.Name}}/bin/{{.Name}}
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
`
)
