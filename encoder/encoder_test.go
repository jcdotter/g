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
	"fmt"
	"strings"
	"testing"

	"github.com/jcdotter/go/test"
)

var config = &test.Config{
	Trace:   true,
	Detail:  true,
	Require: true,
	Msg:     "%s",
}

func TestJson(t *testing.T) {
	var Struct = struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"address"`
	}{
		Name: "James",
		Age:  30,
		Address: struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}{
			City:    "New York",
			Country: "USA",
		},
	}
	fmt.Println(Yaml.Encode(Struct))
}

func TestPrintJson(t *testing.T) {
	gt := test.New(t, config)
	gt.Msg = "Testing Json.Encode(Type: %s\n | Value:\t%s)"
	err := "%!v(PANIC=String method: runtime error: invalid memory address or nil pointer dereference)"
	for n, v := range test.Vals {
		s := Json.Encode(v).String()
		gt.False(s == "" || strings.Contains(s, err), n, s)
		fmt.Println(n, s)
	}
}
