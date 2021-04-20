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

// Package secrets is an interface for encrypting and decrypting secrets
package secrets

import "context"

// Secrets encrypts or decrypts arbitrary data. The data should be as small as possible
type Secrets interface {
	// Init initialise options
	Init(...Option) error
	// Options returns the options
	Options() Options
	// Decrypt a value
	Decrypt([]byte, ...DecryptOption) ([]byte, error)
	// Encrypt a value
	Encrypt([]byte, ...EncryptOption) ([]byte, error)
	// String secrets implementation of String
	String() string
}

type Options struct {
	// Key is a symmetric key for encoding
	Key []byte
	// Private key for decoding
	PrivateKey []byte
	// Public key for encoding
	PublicKey []byte
	// Context for other opts
	Context context.Context
}

// Option sets options
type Option func(*Options)

// Key sets the symmetric secret key
func Key(k []byte) Option {
	return func(o *Options) {
		o.Key = make([]byte, len(k))
		copy(o.Key, k)
	}
}

// PublicKey sets the asymmetric Public Key of this codec
func PublicKey(key []byte) Option {
	return func(o *Options) {
		o.PublicKey = make([]byte, len(key))
		copy(o.PublicKey, key)
	}
}

// PrivateKey sets the asymmetric Private Key of this codec
func PrivateKey(key []byte) Option {
	return func(o *Options) {
		o.PrivateKey = make([]byte, len(key))
		copy(o.PrivateKey, key)
	}
}

// DecryptOptions can be passed to Secrets.Decrypt
type DecryptOptions struct {
	SenderPublicKey []byte
}

// DecryptOption sets DecryptOptions
type DecryptOption func(*DecryptOptions)

// SenderPublicKey is the Public Key of the Secrets that encrypted this message
func SenderPublicKey(key []byte) DecryptOption {
	return func(d *DecryptOptions) {
		d.SenderPublicKey = make([]byte, len(key))
		copy(d.SenderPublicKey, key)
	}
}

// EncryptOptions can be passed to Secrets.Encrypt
type EncryptOptions struct {
	RecipientPublicKey []byte
}

// EncryptOption Sets EncryptOptions
type EncryptOption func(*EncryptOptions)

// RecipientPublicKey is the Public Key of the Secrets that will decrypt this message
func RecipientPublicKey(key []byte) EncryptOption {
	return func(e *EncryptOptions) {
		e.RecipientPublicKey = make([]byte, len(key))
		copy(e.RecipientPublicKey, key)
	}
}
