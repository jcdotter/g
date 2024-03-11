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

package ui

import (
	"context"
	"io"
	"net/http"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/parser"
)

// UI Package Architecture

// Process
// 1. Intake user html template
// 2. Build CSS
//   - parse html classes from template
//   - build and return css from classes

// style, component, page, template
// <ComponentName prop1="value" prop2="value" class="Class1 Class2" />
// how to push data to component

// The ui package is responsible for the user interface of the application. It contains the following sub-packages:
// - css: contains the color palette for the user interface
// - components: contains the reusable components for the user interface
// - pages: contains the different pages of the application
// - templates: contains the HTML templates for the different pages of the application
// - ui.go: contains the main function for the user interface

/*
	package main

	import (
		"net/http"
		"github.com/jcdotter/go/ui"
	)

	func main() {
		http.Handle("/", ui.Page(...))
		http.ListenAndServe(":3000", nil)
	}

	ui.Div(ui.Id("div-001"), ui.Class("container", "mx-auto").Content(
		ui.P(ui.Class("text-center").Content(
			ui.Text("Hello, World!"),
		)),
	))


*/

/*
ATTRIBUTES
- id
- class
- style
- data-*
- aria-*
*/
type Class interface {
	Name() string
	Render(ctx context.Context, w io.Writer) error
}

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

type Page interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type elem struct {
	tag      string            // tag name
	attrs    map[string]string // attributes with values
	flags    []string          // boolean attributes
	contents []Component       // child elements
}

func (c elem) Attrs(attrs map[string]string) {
	for k, v := range attrs {
		c.attrs[k] = v
	}
}

func (c *elem) Flags(flags ...string) {
	c.flags = append(c.flags, flags...)
}

func (c *elem) Content(contents ...Component) {
	c.contents = append(c.contents, contents...)
}

func (c elem) buffer(ctx context.Context, b *buffer.Buffer) (err error) {
	// TODO: add depth for pretty printing
	if err = b.WriteByte('<'); err != nil {
		return
	}
	b.MustWriteString(c.tag)
	for k, v := range c.attrs {
		b.MustWriteByte(' ').
			MustWriteString(k).
			MustWriteByte('=').
			MustWriteByte('"').
			MustWriteString(v).
			MustWriteByte('"')
	}
	b.MustWriteByte('>')
	for _, content := range c.contents {
		if err = content.Render(ctx, b); err != nil {
			return
		}
	}
	b.MustWriteString("</").
		MustWriteString(c.tag).
		MustWriteByte('>')
	return
}

func (c elem) Render(ctx context.Context, w io.Writer) (err error) {
	if b, ok := w.(*buffer.Buffer); ok {
		return c.buffer(ctx, b)
	}
	b := buffer.Pool.Get()
	defer b.Free()
	if err := c.buffer(ctx, b); err != nil {
		return err
	}
	_, err = w.Write(b.Buffer())
	return err
}

type text string

func (t text) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(t))
	return err
}

func Attrs(keyvals ...string) map[string]string {
	attrs := make(map[string]string)
	for i := 0; i < len(keyvals); i += 2 {
		attrs[keyvals[i]] = keyvals[i+1]
	}
	return attrs
}

var attrKeyEnd = parser.OR(parser.NOT(parser.IsChar), parser.Cond('=', '='))

func ParseAttrs(attrs string) (kv map[string]string, f []string) {

	// set up parser
	kv = make(map[string]string)
	f = make([]string, 0, 4)
	var k string
	var v []byte
	var a = []byte(attrs)
	var found bool
	var pos int

	// parse attributes
	for i := 0; i < len(a); i++ {

		// skip whitespace
		if found, i = parser.Next(parser.IsChar, a, i); !found {
			break
		}

		// parse key
		found, pos = parser.Next(attrKeyEnd, a, i)
		if !found {
			f = append(f, string(a[i:]))
			break
		} else if a[pos] == ' ' {
			f = append(f, string(a[i:pos]))
			i = pos
			continue
		}
		k = string(a[i:pos])
		pos++

		// parse value
		v, i = parser.StringItem.Parse(a, pos)
		kv[k] = string(v)

	}
	return
}

func Text(contents string) Component {
	return text(contents)
}

func P(attrs string) Component {
	// TODO:
	// move parse to elem method
	// expand Component interface methods
	// add other html attrs for actions and htmx
	// build out standard html elements
	// add css rendering
	// think about how to include js
	// build out components
	// capabilities for custom components
	// pages
	a, f := ParseAttrs(attrs)
	return elem{tag: "p", attrs: a, flags: f}
}

func Comment() {
}
