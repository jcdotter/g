// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/grpg/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto

import (
	"testing"

	"github.com/jcdotter/gtest"
)

var config = &gtest.Config{
	//PrintTest:   true,
	PrintFail:   true,
	PrintTrace:  true,
	PrintDetail: true,
	FailFatal:   true,
	Msg:         "%s",
}

func TestKey(t *testing.T) {
	gt := gtest.New(t, config)
	gt.Msg = "Crypto Key %s"
	key := NewKey(128)
	gt.Equal(16, len(key), "Length")
}

func TestGCM(t *testing.T) {
	gt := gtest.New(t, config)
	gt.Msg = "Crypto GCM %s"
	key := NewKey(128)
	gcm, e := NewGCM(key)
	gt.Equal(nil, e, "Error")
	gt.NotEqual(nil, gcm, "Value")
}

func TestEncrypt(t *testing.T) {
	gt := gtest.New(t, config)
	gt.Msg = "Crypto Encrypt %s"
	key := NewKey(128)
	text := []byte("hello world!")
	ciphertext, e := Encrypt(key, text)
	gt.Equal(nil, e, "Error")
	gt.NotEqual(nil, ciphertext, "Value")
	utext, e := Decrypt(key, ciphertext)
	gt.Equal(nil, e, "Error")
	gt.Equal(text, utext, "Value")
}
