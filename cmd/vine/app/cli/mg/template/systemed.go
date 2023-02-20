package template

var (
	SystemedSRV = `[Unit]
Description={{.Alias}}
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root

Restart=on-failure
LimitNOFILE=65536
RestartSec=10
startLimitIntervalSec=60

WorkingDirectory=/opt/vine/{{.Name}}
ExecStart=/opt/vine/{{.Name}}/bin/{{.Name}}

PermissionsStartOnly=true
ExecStartPre=

[Install]
WantedBy=multi-user.target
`
)
