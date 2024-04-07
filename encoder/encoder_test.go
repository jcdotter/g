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

package encoder

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	str "github.com/jcdotter/go/strings"
	"github.com/jcdotter/go/test"
	"github.com/jcdotter/go/typ"
)

var config = &test.Config{
	Trace:   true,
	Detail:  true,
	Require: true,
	Msg:     "%s",
}

func TestEncoder(t *testing.T) {
	var Struct = struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"address"`
	}{
		Name: "John Doe",
		Age:  30,
		Address: struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}{
			City:    "New York",
			Country: "USA",
		},
	}

	json := `{"Name":"John Doe","Age":30,"Address":{"City":"New York","Country":"USA"}}`
	yaml := "Name: \"John Doe\"\nAge: 30\nAddress: \n  City: \"New York\"\n  Country: USA"

	jsonResult := Json.Encode(Struct).String()
	yamlResult := Yaml.Encode(Struct).String()

	test.Assert(t, jsonResult, json, "Json.Encode")
	test.Assert(t, yamlResult, yaml, "Yaml.Encode")
}

func TestPrint(t *testing.T) {
	gt := test.New(t, config)
	gt.Msg = "Testing Json.Encode(Type: %s\n | Value:\t%s)"
	err := "%!v(PANIC=String method: runtime error: invalid memory address or nil pointer dereference)"
	for n, v := range test.ValMap(true, 1, "true") {
		s := Json.Encode(v).String()
		gt.False(s == "" || strings.Contains(s, err), n, s)
	}
}

func TestEncodeDecode(t *testing.T) {
	gt := test.New(t, config)

	// Test Json
	gt.Msg = "Testing Json.Encode().Decode(Type: %s\n | Value:\t%s)"
	for n, v := range test.ValMap(true, 1, "true") {
		b := Json.Encode(v).Bytes()
		Json.Reset()
		r := Json.Decode(b).Bytes()
		gt.Equal(b, r, n, v)
	}

	// Test Yaml
	gt.Msg = "Testing Yaml.Encode().Decode(Type: %s\n | Value:\t%s)"
	for n, v := range test.ValMap(true, 1, "true") {
		b := Yaml.Encode(v).Bytes()
		Yaml.Reset()
		r := Yaml.Decode(b).Bytes()
		gt.Equal(b, r, n, v)
	}
}

func BenchmarkMarshaller(b *testing.B) {
	var s string
	var j []byte
	var k, v = typ.MapOf(test.ValMap(true, 1, "true")).KeyVals()
	for i, val := range v {
		key := k[i]
		b.Run(str.Width(key, 35)+"-go-marsh", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				j, _ = json.Marshal(val)
				var b bytes.Buffer
				json.Indent(&b, j, "", "  ")
			}
		})
		b.Run(str.Width(key, 35)+"-encoder", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s = Json.Encode(val).String()
			}
		})
		p := typ.TypeOf(val).PtrType().New().Interface()
		b.Run(str.Width(key, 35)+"-go-unmarsh", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				json.Unmarshal(j, p)
			}
		})
		b.Run(str.Width(key, 35)+"-decoder", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Json.Decode([]byte(s))
			}
		})
	}
}
