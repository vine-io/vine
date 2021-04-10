// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package helper

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lack-io/cli"

	"github.com/lack-io/vine/util/context/metadata"
)

func RequestToContext(c *fiber.Ctx) context.Context {
	ctx := context.Background()
	md := make(metadata.Metadata)
	c.Request().Header.VisitAll(func(key, value []byte) {
		md[string(key)] = string(value)
	})
	return metadata.NewContext(ctx, md)
}

func TLSConfig(ctx *cli.Context) (*tls.Config, error) {
	cert := ctx.String("tls-cert-file")
	key := ctx.String("tls-key-file")
	ca := ctx.String("tls-client-ca-file")

	if len(cert) > 0 && len(key) > 0 {
		certs, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		if len(ca) > 0 {
			caCert, err := ioutil.ReadFile(ca)
			if err != nil {
				return nil, err
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			return &tls.Config{
				Certificates: []tls.Certificate{certs},
				ClientCAs:    caCertPool,
				ClientAuth:   tls.RequireAndVerifyClientCert,
				NextProtos:   []string{"h2", "http/1.1"},
			}, nil
		}

		return &tls.Config{
			Certificates: []tls.Certificate{certs}, NextProtos: []string{"h2", "http/1.1"},
		}, nil
	}

	return nil, errors.New("TLS certificate and key files not specified")
}

// UnexpectedSubcommand checks for erroneous subcommands and prints help and returns error
func UnexpectedSubcommand(ctx *cli.Context) error {
	if first := ctx.Args().First(); first != "" {
		// received something that isn't a subcommand
		return fmt.Errorf("Unrecognized subcommand for %s: %s. Please refer to '%s help'", ctx.App.Name, first, ctx.App.Name)
	}
	return nil
}

func UnexpectedCommand(ctx *cli.Context) error {
	commandName := ""
	// We fall back to os.Args as ctx does not seem to have the original command.
	for _, arg := range os.Args[1:] {
		// Exclude flags
		if !strings.HasPrefix(arg, "-") {
			commandName = arg
		}
	}
	return fmt.Errorf("Unrecognized vine command: %s. Please refer to 'vine help'", commandName)
}

func MissingCommand(ctx *cli.Context) error {
	return fmt.Errorf("No command provided to vine. Please refer to 'vine help'")
}
