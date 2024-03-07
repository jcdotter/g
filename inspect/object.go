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

package inspect

import (
	"github.com/jcdotter/go/data"
)

type Object interface{}

// Array represents a Go array type.
type Array struct {
	typ  *Type
	elem *Type
	len  int
}

func (a *Array) Type() *Type {
	return a.typ
}

func (a *Array) Elem() *Type {
	return a.elem
}

func (a *Array) Len() int {
	return a.len
}

// Chan represents a Go channel type.
type Chan struct {
	typ  *Type
	elem *Type
	dir  byte
}

func (c *Chan) Type() *Type {
	return c.typ
}

func (c *Chan) Elem() *Type {
	return c.elem
}

func (c *Chan) Dir() byte {
	return c.dir
}

// Interface represents a Go interface type.
type Interface struct {
	typ     *Type
	methods *data.Data
}

func (i *Interface) Type() *Type {
	return i.typ
}

func (i *Interface) Methods() *data.Data {
	return i.methods
}

// Pointer represents a Go pointer type.
type Pointer struct{ typ, elem *Type }

func (p *Pointer) Type() *Type {
	return p.typ
}

func (p *Pointer) Elem() *Type {
	return p.elem
}

// Map represents a Go map type.
type Map struct{ typ, key, elem *Type }

func (m *Map) Type() *Type {
	return m.typ
}

func (m *Map) Key() *Type {
	return m.key
}

func (m *Map) Elem() *Type {
	return m.elem
}

// --------------------------------------------------------------------
// STRUCT OBJECT

// Struct represents a Go struct type.
type Struct struct {
	typ    *Type
	fields *data.Data
}

func NewStruct(typ *Type) *Struct {
	return &Struct{
		typ:    typ,
		fields: data.Make[*Field](data.Cap),
	}
}

// Fields returns the fields of the struct,
// excluding fields with a func type.
func (s *Struct) Fields() (fields *data.Data) {
	fields = data.Make[*Field](s.fields.Len())
	for _, f := range s.fields.List() {
		field := f.(*Field)
		if field.typ.Kind() != FUNC {
			fields.Add(field)
		}
	}
	return
}

// Field returns the field with the given name.
func (s *Struct) Field(name string) (field *Field) {
	return s.fields.Get(name).(*Field)
}

// Index returns the field at the given index.
func (s *Struct) Index(i int) (field *Field) {
	return s.fields.Index(i).(*Field)
}

// Funcs returns the fields of the struct with a func type.
func (s *Struct) Funcs() (fields *data.Data) {
	fields = data.Make[*Field](s.fields.Len())
	for _, f := range s.fields.List() {
		field := f.(*Field)
		if field.typ.Kind() == FUNC {
			fields.Add(field)
		}
	}
	return
}

// Methods returns the methods of the struct.
func (s *Struct) Methods() (methods *data.Data) {
	return s.typ.methods
}

// Field represents a Go struct field.
type Field struct {
	typ    *Type
	of     *Type
	name   string
	tag    string
	offset int
}

func (f *Field) Key() string {
	return f.name
}

func (f *Field) Type() *Type {
	return f.typ
}

func (f *Field) Of() *Type {
	return f.of
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) Tag() string {
	return f.tag
}

func (f *Field) Offset() int {
	return f.offset
}
