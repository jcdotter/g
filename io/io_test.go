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

package io

import (
	"os"
	"testing"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/test"
)

var config = &test.Config{
	PrintTest:   false,
	PrintFail:   true,
	PrintTrace:  true,
	PrintDetail: true,
	FailFatal:   true,
	Msg:         "%s",
}

func TestNew(t *testing.T) {
	gt := test.New(t, config)

	var out = buffer.Pool.Get()
	var io = New()
	gt.NotEqual(nil, io, "io != nil")
	gt.NotEqual(nil, io.Buffer(), "io buffer != nil")
	gt.Equal(os.Stdin, io.Out(), "io out == os.Stdin")

	io.SetOut(out)
	gt.Equal(out, io.Out(), "io out == out")

	data := struct {
		S1 string
		S2 string
	}{"This", "works"}
	io.AppendString("{{ .S1 }} {{ .S2 }}!")
	io.Inject(data)
	io.Output()
	gt.Equal("This works!", out.String(), "out == 'This works!'")

}
