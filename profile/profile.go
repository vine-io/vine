// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package profile is for specific profiles
// @todo this package is the definition of cruft and
// should be rewritten in a more elegant way
package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine/service/auth/jwt"
	"github.com/lack-io/vine/service/auth/noop"
	"github.com/lack-io/vine/service/broker"
	memBroker "github.com/lack-io/vine/service/broker/memory"
	"github.com/lack-io/vine/service/build/golang"
	"github.com/lack-io/vine/service/client"
	"github.com/lack-io/vine/service/config"
	storeConfig "github.com/lack-io/vine/service/config/store"
	evStore "github.com/lack-io/vine/service/events/store"
	memStream "github.com/lack-io/vine/service/events/stream/memory"
	"github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/registry"
	"github.com/lack-io/vine/service/registry/mdns"
	"github.com/lack-io/vine/service/registry/memory"
	"github.com/lack-io/vine/service/router"
	k8sRouter "github.com/lack-io/vine/service/router/kubernetes"
	regRouter "github.com/lack-io/vine/service/router/registry"
	"github.com/lack-io/vine/service/runtime/kubernetes"
	"github.com/lack-io/vine/service/runtime/local"
	"github.com/lack-io/vine/service/server"
	"github.com/lack-io/vine/service/store/file"
	mem "github.com/lack-io/vine/service/store/memory"

	inAuth "github.com/lack-io/vine/internal/auth"
	"github.com/lack-io/vine/internal/user"
	vineAuth "github.com/lack-io/vine/service/auth"
	vineBuilder "github.com/lack-io/vine/service/build"
	vineEvents "github.com/lack-io/vine/service/events"
	vineRegistry "github.com/lack-io/vine/service/registry"
	vineRouter "github.com/lack-io/vine/service/router"
	vineRuntime "github.com/lack-io/vine/service/runtime"
	vineStore "github.com/lack-io/vine/service/store"
)

// profiles which when called will configure vine to run in that environment
var profiles = map[string]*Profile{
	// built in profiles
	"client":     Client,
	"service":    Service,
	"test":       Test,
	"local":      Local,
	"kubernetes": Kubernetes,
}

// Profile configures an environment
type Profile struct {
	// name of the profile
	Name string
	// function used for setup
	Setup func(*cli.Context) error
	// TODO: presetup dependencies
	// e.g start resources
}

// Register a profile
func Register(name string, p *Profile) error {
	if _, ok := profiles[name]; ok {
		return fmt.Errorf("profile %s already exists", name)
	}
	profiles[name] = p
	return nil
}

// Load a profile
func Load(name string) (*Profile, error) {
	v, ok := profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %s does not exist", name)
	}
	return v, nil
}

// Client profile is for any entrypoint that behaves as a client
var Client = &Profile{
	Name:  "client",
	Setup: func(ctx *cli.Context) error { return nil },
}

// Local profile to run locally
var Local = &Profile{
	Name: "local",
	Setup: func(ctx *cli.Context) error {
		vineAuth.DefaultAuth = jwt.NewAuth()
		vineStore.DefaultStore = file.NewStore(file.WithDir(filepath.Join(user.Dir, "server", "store")))
		SetupConfigSecretKey(ctx)
		config.DefaultConfig, _ = storeConfig.NewConfig(vineStore.DefaultStore, "")
		SetupBroker(memBroker.NewBroker())
		SetupRegistry(mdns.NewRegistry())
		SetupJWT(ctx)

		// use the local runtime, note: the local runtime is designed to run source code directly so
		// the runtime builder should NOT be set when using this implementation
		vineRuntime.DefaultRuntime = local.NewRuntime()

		var err error
		vineEvents.DefaultStream, err = memStream.NewStream()
		if err != nil {
			logger.Fatalf("Error configuring stream: %v", err)
		}
		vineEvents.DefaultStore = evStore.NewStore(
			evStore.WithStore(vineStore.DefaultStore),
		)

		vineStore.DefaultBlobStore, err = file.NewBlobStore()
		if err != nil {
			logger.Fatalf("Error configuring file blob store: %v", err)
		}

		return nil
	},
}

// Kubernetes profile to run on kubernetes with zero deps. Designed for use with the vine helm chart
var Kubernetes = &Profile{
	Name: "kubernetes",
	Setup: func(ctx *cli.Context) (err error) {
		vineAuth.DefaultAuth = jwt.NewAuth()
		SetupJWT(ctx)

		vineRuntime.DefaultRuntime = kubernetes.NewRuntime()
		vineBuilder.DefaultBuilder, err = golang.NewBuilder()
		if err != nil {
			logger.Fatalf("Error configuring golang builder: %v", err)
		}

		vineEvents.DefaultStream, err = memStream.NewStream()
		if err != nil {
			logger.Fatalf("Error configuring stream: %v", err)
		}

		vineStore.DefaultStore = file.NewStore(file.WithDir("/store"))
		vineStore.DefaultBlobStore, err = file.NewBlobStore(file.WithDir("/store/blob"))
		if err != nil {
			logger.Fatalf("Error configuring file blob store: %v", err)
		}

		// the registry service uses the memory registry, the other core services will use the default
		// rpc client and call the registry service
		if ctx.Args().Get(1) == "registry" {
			SetupRegistry(memory.NewRegistry())
		}

		// the broker service uses the memory broker, the other core services will use the default
		// rpc client and call the broker service
		if ctx.Args().Get(1) == "broker" {
			SetupBroker(memBroker.NewBroker())
		}

		config.DefaultConfig, err = storeConfig.NewConfig(vineStore.DefaultStore, "")
		if err != nil {
			logger.Fatalf("Error configuring config: %v", err)
		}
		SetupConfigSecretKey(ctx)

		vineRouter.DefaultRouter = k8sRouter.NewRouter()
		client.DefaultClient.Init(client.Router(vineRouter.DefaultRouter))
		return nil
	},
}

// Service is the default for any services run
var Service = &Profile{
	Name:  "service",
	Setup: func(ctx *cli.Context) error { return nil },
}

// Test profile is used for the go test suite
var Test = &Profile{
	Name: "test",
	Setup: func(ctx *cli.Context) error {
		vineAuth.DefaultAuth = noop.NewAuth()
		vineStore.DefaultStore = mem.NewStore()
		vineStore.DefaultBlobStore, _ = file.NewBlobStore()
		config.DefaultConfig, _ = storeConfig.NewConfig(vineStore.DefaultStore, "")
		SetupRegistry(memory.NewRegistry())
		return nil
	},
}

// SetupRegistry configures the registry
func SetupRegistry(reg registry.Registry) {
	vineRegistry.DefaultRegistry = reg
	vineRouter.DefaultRouter = regRouter.NewRouter(router.Registry(reg))
	client.DefaultClient.Init(client.Registry(reg))
	server.DefaultServer.Init(server.Registry(reg))
}

// SetupBroker configures the broker
func SetupBroker(b broker.Broker) {
	broker.DefaultBroker = b
	client.DefaultClient.Init(client.Broker(b))
	server.DefaultServer.Init(server.Broker(b))
}

// SetupJWT configures the default internal system rules
func SetupJWT(ctx *cli.Context) {
	for _, rule := range inAuth.SystemRules {
		if err := vineAuth.DefaultAuth.Grant(rule); err != nil {
			logger.Fatal("Error creating default rule: %v", err)
		}
	}
}

func SetupConfigSecretKey(ctx *cli.Context) {
	key := ctx.String("config_secret_key")
	if len(key) == 0 {
		k, err := user.GetConfigSecretKey()
		if err != nil {
			logger.Fatal("Error getting config secret: %v", err)
		}
		os.Setenv("VINE_CONFIG_SECRET_KEY", k)
	}
}
