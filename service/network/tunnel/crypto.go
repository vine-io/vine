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

package tunnel

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"

	"github.com/oxtoacart/bpool"
)

var (
	// the local buffer pool
	// gcmStandardNonceSize from crypto/cipher/gcm.go is 12 bytes
	// 100 - is max size of pool
	noncePool = bpool.NewBytePool(100, 12)
)

// hash hashes the data into 32 bytes key and returns it
// hash uses sha256 underneath to hash the supplied key
func hash(key []byte) []byte {
	sum := sha256.Sum256(key)
	return sum[:]
}

// Encrypt encrypts data and returns the encrypted data
func Encrypt(gcm cipher.AEAD, data []byte) ([]byte, error) {
	var err error

	// get new byte array the size of the nonce from pool
	// NOTE: we might use smaller nonce size in the future
	nonce := noncePool.Get()
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	defer noncePool.Put(nonce)

	// NOTE: we prepend the nonce to the payload
	// we need to do this as we need the same nonce
	// to decrypt the payload when receiving it
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt decrypts the payload and returns the decrypted data
func newCipher(key []byte) (cipher.AEAD, error) {
	var err error

	// generate a new AES cipher using our 32 byte key for decrypting the message
	c, err := aes.NewCipher(hash(key))
	if err != nil {
		return nil, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func Decrypt(gcm cipher.AEAD, data []byte) ([]byte, error) {
	var err error

	nonceSize := gcm.NonceSize()

	if len(data) < nonceSize {
		return nil, ErrDecryptingData
	}

	// NOTE: we need to parse out nonce from the payload
	// we prepend the nonce to every encrypted payload
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	ciphertext, err = gcm.Open(ciphertext[:0], nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}
