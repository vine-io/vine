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

// Package box is an asymmetric implementation of config/secrets using nacl/box
package box

import (
	"crypto/rand"
	"errors"
	"fmt"

	naclbox "golang.org/x/crypto/nacl/box"

	"github.com/lack-io/vine/lib/config/secrets"
)

const keyLength = 32

type box struct {
	options secrets.Options

	publicKey  [keyLength]byte
	privateKey [keyLength]byte
}

// NewSecrets returns a nacl-box codec
func NewSecrets(opts ...secrets.Option) secrets.Secrets {
	b := &box{}
	for _, o := range opts {
		o(&b.options)
	}
	return b
}

// Init initialises a box
func (b *box) Init(opts ...secrets.Option) error {
	for _, o := range opts {
		o(&b.options)
	}
	if len(b.options.PrivateKey) != keyLength || len(b.options.PublicKey) != keyLength {
		return fmt.Errorf("a public key and a private key of length %d must both be provided", keyLength)
	}
	copy(b.privateKey[:], b.options.PrivateKey)
	copy(b.publicKey[:], b.options.PublicKey)
	return nil
}

// Options returns options
func (b *box) Options() secrets.Options {
	return b.options
}

// String returns nacl-box
func (*box) String() string {
	return "nacl-box"
}

// Encrypt encrypts a message with the sender's private key and the receipient's public key
func (b *box) Encrypt(in []byte, opts ...secrets.EncryptOption) ([]byte, error) {
	var options secrets.EncryptOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.RecipientPublicKey) != keyLength {
		return []byte{}, errors.New("recepient's public key must be provided")
	}
	var recipientPublicKey [keyLength]byte
	copy(recipientPublicKey[:], options.RecipientPublicKey)
	var nonce [24]byte
	if _, err := rand.Reader.Read(nonce[:]); err != nil {
		return []byte{}, fmt.Errorf("%w couldn't obtain a random nonce from crypto/rand", err)
	}
	return naclbox.Seal(nonce[:], in, &nonce, &recipientPublicKey, &b.privateKey), nil
}

// Decrypt Decrypts a message with the receiver's private key and the sender's public key
func (b *box) Decrypt(in []byte, opts ...secrets.DecryptOption) ([]byte, error) {
	var options secrets.DecryptOptions
	for _, o := range opts {
		o(&options)
	}
	if len(options.SenderPublicKey) != keyLength {
		return []byte{}, errors.New("sender's public key bust be provided")
	}
	var nonce [24]byte
	var senderPublicKey [32]byte
	copy(nonce[:], in[:24])
	copy(senderPublicKey[:], options.SenderPublicKey)
	decrypted, ok := naclbox.Open(nil, in[24:], &nonce, &senderPublicKey, &b.privateKey)
	if !ok {
		return []byte{}, errors.New("incoming message couldn't be verified / decrypted")
	}
	return decrypted, nil
}
