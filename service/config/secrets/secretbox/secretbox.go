// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package secretbox is a config/secrets implementation that uses nacl/secretbox
// to do symmetric encryption / verification
package secretbox

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"

	"github.com/lack-io/vine/service/config/secrets"
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
