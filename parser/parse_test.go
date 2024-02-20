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

package parser

import (
	"testing"

	"github.com/jcdotter/go/test"
)

func TestNum(t *testing.T) {
	var n []byte
	n, _ = Number([]byte("123 "), 0)
	test.Print(t, Int(n), 123, "parse int")
	n, _ = Number([]byte("123.456 "), 0)
	test.Print(t, Float(n), 123.456, "parse float")
	n, _ = Number([]byte("123.456e-2 "), 0)
	test.Print(t, Float(n), 1.23456, "parse exponent")
}

/* func TestString(t *testing.T) {
	var b []byte
	var s string
	b, _ = StringLit([]byte(`"hello\nworld" `), 0)
	test.Printr(t, string(b), `"hello"`, "parse string literal")
	s = String(b)
	test.Printr(t, s, "hello\nworld", "parse string")
} */
