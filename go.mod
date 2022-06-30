module github.com/vine-io/vine

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/evanphx/json-patch/v5 v5.2.0
	github.com/felixge/httpsnoop v1.0.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.7
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/hcl v1.0.0
	github.com/imdario/mergo v0.3.11
	github.com/jinzhu/inflection v1.0.0
	github.com/jinzhu/now v1.1.1
	github.com/json-iterator/go v1.1.12
	github.com/kr/pretty v0.2.1
	github.com/miekg/dns v1.1.50
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e
	github.com/stretchr/testify v1.7.0
	github.com/vine-io/cli v1.3.0
	github.com/vine-io/gscheduler v0.3.0
	github.com/xlab/treeprint v1.0.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	google.golang.org/grpc v1.47.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace google.golang.org/grpc => google.golang.org/grpc v1.47.0
