package template

var (
	SystemedSRV = `[Unit]
Description={{.Alias}}
After=network.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/vine/{{.Name}}
EnvironmentFile=-/opt/vine/{{.Name}}/config/{{.Name}}.ini
ExecStart=/opt/vine/{{.Name}}/bin/{{.Name}} \
  --server-name=${SERVER_NAME} \
  --server-address=${SERVER_ADDRESS} \
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
`
)
