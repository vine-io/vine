package template

var (
	SystemedSRV = `[Unit]
Description={{.FQDN}}
After=network.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/vine/{{.Alias}}
EnvironmentFile=-/opt/vine/{{.Alias}}/config/{{.Alias}}.conf
ExecStart=/opt/vine/{{.Alias}}/bin/{{.Alias}} \
  --server-name=${SERVER_NAME} \
  --server-address=${SERVER_ADDRESS} \
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
`
)
