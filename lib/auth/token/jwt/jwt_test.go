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

package jwt

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/lack-io/vine/lib/auth"
	"github.com/lack-io/vine/lib/auth/token"
)

func TestGenerate(t *testing.T) {
	privKey, err := ioutil.ReadFile("test/sample_key")
	if err != nil {
		t.Fatalf("Unable to read private key: %v", err)
	}

	j := NewTokenProvider(
		token.WithPrivateKey(string(privKey)),
	)

	_, err = j.Generate(&auth.Account{ID: "test"})
	if err != nil {
		t.Fatalf("Generate returned %v error, expected nil", err)
	}
}

func TestInspect(t *testing.T) {
	pubKey, err := ioutil.ReadFile("test/sample_key.pub")
	if err != nil {
		t.Fatalf("Unable to read public key: %v", err)
	}
	privKey, err := ioutil.ReadFile("test/sample_key")
	if err != nil {
		t.Fatalf("Unable to read private key: %v", err)
	}

	j := NewTokenProvider(
		token.WithPublicKey(string(pubKey)),
		token.WithPrivateKey(string(privKey)),
	)

	t.Run("Valid token", func(t *testing.T) {
		md := map[string]string{"foo": "bar"}
		scopes := []string{"admin"}
		subject := "test"

		acc := &auth.Account{ID: subject, Scopes: scopes, Metadata: md}
		tok, err := j.Generate(acc)
		if err != nil {
			t.Fatalf("Generate returned %v error, expected nil", err)
		}

		tok2, err := j.Inspect(tok.Token)
		if err != nil {
			t.Fatalf("Inspect returned %v error, expected nil", err)
		}
		if acc.ID != subject {
			t.Errorf("Inspect returned %v as the token subject, expected %v", acc.ID, subject)
		}
		if len(tok2.Scopes) != len(scopes) {
			t.Errorf("Inspect returned %v scopes, expected %v", len(tok2.Scopes), len(scopes))
		}
		if len(tok2.Metadata) != len(md) {
			t.Errorf("Inspect returned %v as the token metadata, expected %v", tok2.Metadata, md)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		tok, err := j.Generate(&auth.Account{}, token.WithExpiry(-10*time.Second))
		if err != nil {
			t.Fatalf("Generate returned %v error, expected nil", err)
		}

		if _, err = j.Inspect(tok.Token); err != token.ErrInvalidToken {
			t.Fatalf("Inspect returned %v error, expected %v", err, token.ErrInvalidToken)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := j.Inspect("Invalid token")
		if err != token.ErrInvalidToken {
			t.Fatalf("Inspect returned %v error, expected %v", err, token.ErrInvalidToken)
		}
	})

}
