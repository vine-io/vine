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
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := []byte("tokenpassphrase")
	gcm, err := newCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("supersecret")

	cipherText, err := Encrypt(gcm, data)
	if err != nil {
		t.Errorf("failed to encrypt data: %v", err)
	}

	// verify the cipherText is not the same as data
	if bytes.Equal(data, cipherText) {
		t.Error("encrypted data are the same as plaintext")
	}
}

func TestDecrypt(t *testing.T) {
	key := []byte("tokenpassphrase")
	gcm, err := newCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("supersecret")

	cipherText, err := Encrypt(gcm, data)
	if err != nil {
		t.Errorf("failed to encrypt data: %v", err)
	}

	plainText, err := Decrypt(gcm, cipherText)
	if err != nil {
		t.Errorf("failed to decrypt data: %v", err)
	}

	// verify the plainText is the same as data
	if !bytes.Equal(data, plainText) {
		t.Error("decrypted data not the same as plaintext")
	}
}
