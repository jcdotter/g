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

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/parser"
)

// ----------------------------------------------------------------------------
// HTML Element

// Elem represents an HTML element,
// which can be modified and rendered.
// It implements the Component interface.
type Elem struct {
	Tag   string            // tag name
	Attrs map[string]string // attributes with values
	Flags []string          // boolean attributes
	Outer Component         // parent element
	Inner []Component       // child elements
}

// El creates a new HTML element.
func El(tag string, attrs ...string) *Elem {
	e := &Elem{Tag: tag}
	for _, attr := range attrs {
		e.SetAttrs(attr)
	}
	return e
}

// GetAttr returns the value of an attribute.
func (e *Elem) GetAttr(attr string) string {
	if v, ok := e.Attrs[attr]; ok {
		return v
	}
	return ""
}

// GetFlag returns true if the element has a boolean attribute.
func (e *Elem) GetFlag(flag string) bool {
	for _, f := range e.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// Attrs sets the attributes of the element.
func (e *Elem) SetAttrs(attrs string) *Elem {
	e.Attrs, e.Flags = ParseAttrs(attrs)
	return e
}

// SetAttr sets an attribute of the element.
func (e *Elem) SetAttr(attr, value string) *Elem {
	e.Attrs[attr] = value
	return e
}

// SetFlag sets a boolean attribute of the element.
func (e *Elem) SetFlag(flag string) *Elem {
	e.Flags = append(e.Flags, flag)
	return e
}

// Unset removes an attribute from the element.
func (e *Elem) Unset(attr string) *Elem {
	if _, ok := e.Attrs[attr]; ok {
		delete(e.Attrs, attr)
		return e
	}
	for i, flag := range e.Flags {
		if flag == attr {
			e.Flags = append(e.Flags[:i], e.Flags[i+1:]...)
			return e
		}
	}
	return e
}

// SetInner sets the inner contents of the element.
func (e *Elem) SetInner(contents ...Component) *Elem {
	e.Inner = contents
	return e
}

// AppendInner appends contents to the inner contents of the element.
func (e *Elem) AppendInner(contents ...Component) *Elem {
	e.Inner = append(e.Inner, contents...)
	return e
}

// PrependInner prepends contents to the inner contents of the element.
func (e *Elem) PrependInner(contents ...Component) *Elem {
	e.Inner = append(contents, e.Inner...)
	return e
}

// SetOuter wraps the element in the provided component
// and sets the new outer element as the inner of the
// prior outer element.
func (e *Elem) SetOuter(outer *Elem) *Elem {

	// set new outer element as the inner
	// of the prior outer element
	index := -1
	if e.Outer != nil {
		if o, ok := e.Outer.(*Elem); ok {
			for i, inner := range o.Inner {
				if inner == e {
					index = i
					break
				}
			}
			if index > -1 {
				o.Inner[index] = outer
			}
		}
	}

	// set new outer element
	e.Outer = outer
	outer.SetInner(e)
	return e
}

// buffer writes the element to a buffer.
func (e Elem) buffer(ctx context.Context, b *buffer.Buffer) (err error) {
	// TODO: add depth for pretty printing
	// comment tag
	if err = b.WriteByte('<'); err != nil {
		return
	}
	b.MustWriteString(e.Tag)
	for k, v := range e.Attrs {
		b.MustWriteByte(' ').
			MustWriteString(k).
			MustWriteByte('=').
			MustWriteByte('"').
			MustWriteString(v).
			MustWriteByte('"')
	}
	for _, f := range e.Flags {
		b.MustWriteByte(' ').
			MustWriteString(f)
	}
	if len(e.Inner) == 0 {
		b.MustWriteString("/>")
		return
	}
	b.MustWriteByte('>')
	for _, content := range e.Inner {
		if err = content.Render(ctx, b); err != nil {
			return
		}
	}
	b.MustWriteString("</").
		MustWriteString(e.Tag).
		MustWriteByte('>')
	return
}

func (e Elem) Render(ctx context.Context, w io.Writer) (err error) {
	if b, ok := w.(*buffer.Buffer); ok {
		return e.buffer(ctx, b)
	}
	b := buffer.Pool.Get()
	defer b.Free()
	if err := e.buffer(ctx, b); err != nil {
		return err
	}
	_, err = w.Write(b.Buffer())
	return err
}

// ----------------------------------------------------------------------------
// Element Helpers

// attrKeyEnd is a parser condition for
// the end of an attribute key.
var attrKeyEnd = parser.OR(
	parser.NOT(parser.IsChar),
	parser.Cond('=', '='),
)

// ParseAttrs parses a string of HTML
// attributes and returns a map of key-value
// pairs and a slice of boolean flags. The
// flags are attributes without values.
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
