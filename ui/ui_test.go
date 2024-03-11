// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License";
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

package ui

import (
	"context"
	"testing"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/test"
)

func TestParseAttrs(t *testing.T) {
	attrs := `id="div-001" class="container mx-auto" visible`
	a, f := ParseAttrs(attrs)
	test.Assert(t, a["id"], "div-001", "ParseAttrs.id")
	test.Assert(t, a["class"], "container mx-auto", "ParseAttrs.class")
	test.Assert(t, "visible", f[0], "ParseAttrs.visible")
}

func TestElem(t *testing.T) {
	e := El(`div`, `id="div-001" class="container mx-auto" visible`).SetInner(
		El(`p`, `class="text-center"`).SetInner(
			Text("Hello, World!"),
		),
	)
	test.Assert(t, e.Tag, "div", "Elem.Tag")
	test.Assert(t, e.GetAttr("id"), "div-001", `Elem.GetAttr("id")`)
	test.Assert(t, e.GetAttr("class"), "container mx-auto", `Elem.GetAttr("class")`)
	test.Assert(t, true, e.GetFlag("visible"), `Elem.GetFlag("visible")`)
	test.Assert(t, e.Inner[0].(*Elem).Tag, "p", "Elem.Inner[0].Tag")
	test.Assert(t, e.Inner[0].(*Elem).GetAttr("class"), "text-center", `Elem.Inner[0].GetAttr("class")`)
	test.Assert(t, string(e.Inner[0].(*Elem).Inner[0].(text)), "Hello, World!", `Elem.Inner[0].Inner[0].Text`)
}

func TestRenderElem(t *testing.T) {
	e := El(`div`, `id="div-001" class="container mx-auto" visible`).SetInner(
		El(`p`, `class="text-center"`).SetInner(
			Text("Hello, World!"),
		),
	)
	b := buffer.New()
	err := e.Render(context.Background(), b)
	test.Assert(t, err, nil, "RenderElem.err")
	test.Assert(t, b.String(), `<div id="div-001" class="container mx-auto" visible><p class="text-center">Hello, World!</p></div>`, "RenderElem")
}
