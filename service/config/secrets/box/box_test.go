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

package box

import (
	"crypto/rand"
	"reflect"
	"testing"

	naclbox "golang.org/x/crypto/nacl/box"

	"github.com/lack-io/vine/service/config/secrets"
)

func TestBox(t *testing.T) {
	alicePublicKey, alicePrivateKey, err := naclbox.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	bobPublicKey, bobPrivateKey, err := naclbox.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	alice, bob := NewSecrets(secrets.PublicKey(alicePublicKey[:]), secrets.PrivateKey(alicePrivateKey[:])), NewSecrets()
	if err := alice.Init(); err != nil {
		t.Error(err)
	}
	if err := bob.Init(secrets.PublicKey(bobPublicKey[:]), secrets.PrivateKey(bobPrivateKey[:])); err != nil {
		t.Error(err)
	}
	if alice.String() != "nacl-box" {
		t.Error("String() doesn't return nacl-box")
	}
	aliceSecret := []byte("Why is a raven like a writing-desk?")
	if _, err := alice.Encrypt(aliceSecret); err == nil {
		t.Error("alice.Encrypt succeded without a public key")
	}
	enc, err := alice.Encrypt(aliceSecret, secrets.RecipientPublicKey(bob.Options().PublicKey))
	if err != nil {
		t.Error("alice.Encrypt failed")
	}
	if _, err := bob.Decrypt(enc); err == nil {
		t.Error("bob.Decrypt succeded without a public key")
	}
	if dec, err := bob.Decrypt(enc, secrets.SenderPublicKey(alice.Options().PublicKey)); err == nil {
		if !reflect.DeepEqual(dec, aliceSecret) {
			t.Errorf("Bob's decrypted message didn't match Alice's encrypted message: %v != %v", aliceSecret, dec)
		}
	} else {
		t.Errorf("bob.Decrypt failed (%s)", err)
	}

	bobSecret := []byte("I haven't the slightest idea")
	enc, err = bob.Encrypt(bobSecret, secrets.RecipientPublicKey(alice.Options().PublicKey))
	if err != nil {
		t.Error(err)
	}
	dec, err := alice.Decrypt(enc, secrets.SenderPublicKey(bob.Options().PrivateKey))
	if err == nil {
		t.Error(err)
	}
	dec, err = alice.Decrypt(enc, secrets.SenderPublicKey(bob.Options().PublicKey))
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dec, bobSecret) {
		t.Errorf("Alice's decrypted message didn't match Bob's encrypted message %v != %v", bobSecret, dec)
	}
}
