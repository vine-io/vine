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

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/lack-io/cli"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/lack-io/vine/client/cli/namespace"
	"github.com/lack-io/vine/client/cli/signup"
	"github.com/lack-io/vine/client/cli/token"
	"github.com/lack-io/vine/client/cli/util"
	"github.com/lack-io/vine/internal/report"
	"github.com/lack-io/vine/service/auth"
)

// login flow.
func login(ctx *cli.Context) error {
	// assuming --otp go to platform.Signup
	if isOTP := ctx.Bool("otp"); isOTP {
		return signup.Run(ctx)
	}

	// otherwise assume username/password login

	// get the environment
	env, err := util.GetEnv(ctx)
	if err != nil {
		return err
	}
	// get the username
	username := ctx.String("username")

	// username is blank
	if len(username) == 0 {
		fmt.Print("Enter username: ")
		// read out the username from prompt if blank
		reader := bufio.NewReader(os.Stdin)
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	}

	ns, err := namespace.Get(env.Name)
	if err != nil {
		return err
	}

	// clear tokens and try again
	if err := token.Remove(ctx); err != nil {
		report.Errorf(ctx, "%v: Token remove: %v", username, err.Error())
		return err
	}

	password := ctx.String("password")
	if len(password) == 0 {
		pw, err := getPassword()
		if err != nil {
			return err
		}
		password = pw
		fmt.Println()
	}
	tok, err := auth.Token(auth.WithCredentials(username, password), auth.WithTokenIssuer(ns))
	if err != nil {
		report.Errorf(ctx, "%v: Getting token: %v", username, err.Error())
		return err
	}
	token.Save(ctx, tok)

	fmt.Println("Successfully logged in.")
	return nil
}

// taken from https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
func getPassword() (string, error) {
	fmt.Print("Enter password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}

func logout(ctx *cli.Context) error {
	return token.Remove(ctx)
}
