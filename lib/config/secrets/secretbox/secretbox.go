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

// Package secretbox is a config/secrets implementation that uses nacl/secretbox
// to do symmetric encryption / verification
package secretbox

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"

	"github.com/vine-io/vine/lib/config/secrets"
)

const keyLength = 32

type secretBox struct {
	options secrets.Options

	secretKey [keyLength]byte
}

// NewSecrets returns a secretbox codec
func NewSecrets(opts ...secrets.Option) secrets.Secrets {
	sb := &secretBox{}
	for _, o := range opts {
		o(&sb.options)
	}
	return sb
}

func (s *secretBox) Init(opts ...secrets.Option) error {
	for _, o := range opts {
		o(&s.options)
	}
	if len(s.options.Key) == 0 {
		return errors.New("no secret key is defined")
	}
	if len(s.options.Key) != keyLength {
		return fmt.Errorf("secret key must be %d bytes long", keyLength)
	}
	copy(s.secretKey[:], s.options.Key)
	return nil
}

func (s *secretBox) Options() secrets.Options {
	return s.options
}

func (s *secretBox) String() string {
	return "nacl-secretbox"
}

func (s *secretBox) Encrypt(in []byte, opts ...secrets.EncryptOption) ([]byte, error) {
	// no opts are expected, so they are ignored

	// there must be a unique nonce for each message
	var nonce [24]byte
	if _, err := rand.Reader.Read(nonce[:]); err != nil {
		return []byte{}, fmt.Errorf("%w couldn't obtain a random nonce from crypto/rand", err)
	}
	return secretbox.Seal(nonce[:], in, &nonce, &s.secretKey), nil
}

func (s *secretBox) Decrypt(in []byte, opts ...secrets.DecryptOption) ([]byte, error) {
	// no options are expected, so they are ignored

	var decryptNonce [24]byte
	copy(decryptNonce[:], in[:24])
	decrypted, ok := secretbox.Open(nil, in[24:], &decryptNonce, &s.secretKey)
	if !ok {
		return []byte{}, errors.New("decryption failed (is the key set correctly?)")
	}
	return decrypted, nil
}
