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

package basic

import (
	"testing"

	"github.com/lack-io/vine/lib/auth"
	"github.com/lack-io/vine/lib/auth/token"
)

func TestGenerate(t *testing.T) {
	//store := memory.NewStore()
	//b := NewTokenProvider(token.WithStore(store))
	b := NewTokenProvider()

	_, err := b.Generate(&auth.Account{ID: "test"})
	if err != nil {
		t.Fatalf("Generate returned %v error, expected nil", err)
	}

	//recs, err := store.List()
	//if err != nil {
	//	t.Fatalf("Unable to read from store: %v", err)
	//}
	//if len(recs) != 1 {
	//	t.Errorf("Generate didn't write to the store, expected 1 record, got %v", len(recs))
	//}
}

func TestInspect(t *testing.T) {
	//store := memory.NewStore()
	//b := NewTokenProvider(token.WithStore(store))
	b := NewTokenProvider()

	t.Run("Valid token", func(t *testing.T) {
		md := map[string]string{"foo": "bar"}
		scopes := []string{"admin"}
		subject := "test"

		tok, err := b.Generate(&auth.Account{ID: subject, Scopes: scopes, Metadata: md})
		if err != nil {
			t.Fatalf("Generate returned %v error, expected nil", err)
		}

		tok2, err := b.Inspect(tok.Token)
		if err != nil {
			t.Fatalf("Inspect returned %v error, expected nil", err)
		}
		if tok.ID != subject {
			t.Errorf("Inspect returned %v as the token subject, expected %v", tok2.ID, subject)
		}
		if len(tok2.Scopes) != len(scopes) {
			t.Errorf("Inspect returned %v scopes, expected %v", len(tok2.Scopes), len(scopes))
		}
		if len(tok2.Metadata) != len(md) {
			t.Errorf("Inspect returned %v as the token metadata, expected %v", tok2.Metadata, md)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := b.Inspect("Invalid token")
		if err != token.ErrInvalidToken {
			t.Fatalf("Inspect returned %v error, expected %v", err, token.ErrInvalidToken)
		}
	})
}
