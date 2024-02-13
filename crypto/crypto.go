// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// NewKey generates a new encryption key of the given bit size
func NewKey(bits int) []byte {
	s := make([]byte, bits/8)
	rand.Read(s)
	return s
}

// NewGCM returns the given 128-bit, block cipher wrapped
// in Galois Counter Mode with the standard nonce length,
// using the given key to generate a new, random nonce
func NewGCM(key []byte) (gcm cipher.AEAD, e error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	gcm, e = cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	return
}

// Encrypt encrypts the given plaintext using the given key
func Encrypt(key, plaintext []byte) (ciphertext []byte, e error) {
	gcm, e := NewGCM(key)
	if e != nil {
		return nil, e
	}
	nonce := make([]byte, gcm.NonceSize())
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts the given ciphertext using the given key
func Decrypt(key, ciphertext []byte) (plaintext []byte, e error) {
	gcm, e := NewGCM(key)
	if e != nil {
		return nil, e
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptString encrypts the given plaintext string using the given key
func EncryptString(key []byte, plaintext string) (ciphertext string, e error) {
	b, e := Encrypt(key, []byte(plaintext))
	return string(b), e
}

// DecryptString decrypts the given ciphertext string using the given key
func DecryptString(key []byte, ciphertext string) (plaintext string, e error) {
	b, e := Decrypt(key, []byte(ciphertext))
	return string(b), e
}
