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

package typ

import (
	"reflect"
	"strconv"
	"unsafe"
)

// ------------------------------------------------------------ /
// Value IMPLEMENTATION
// inspired by golang standard reflect.Value
// with expanded methods and type conversations

// Value contains a pointer to a value and to its datatype
type Value struct{ reflect.Value }

// ValueOf returns a Value of any type
func ValueOf(a any) (v Value) {
	v.Value = reflect.ValueOf(a)
	return
}

// FromReflect returns a Value from a reflect.Value
func FromReflectValue(r reflect.Value) (v Value) {
	v.Value = r
	return
}

// Reflect returns a reflect.Value from a Value
func (v Value) Reflect() reflect.Value {
	return v.Value
}

// Iface returns an interface type converter
func (v Value) Iface() Iface {
	return *(*Iface)(unsafe.Pointer(&v.Value))
}

// Type returns the type of the value
func (v Value) Type() *Type {
	return v.Iface().Type
}

// Pointer returns the pointer to the value
func (v Value) Pointer() unsafe.Pointer {
	return v.Iface().Pointer
}

// Elem returns the value that the pointer points to
func (v Value) Elem() Value {
	return FromReflectValue(v.Value.Elem())
}

// Kind returns the kind of the value
func (v Value) Kind() byte {
	return byte(v.Kind())
}

// KindX returns the extended kind of the value
// to include aliases and special types
func (v Value) KindX() byte {
	return v.Type().KindX()
}

// Slice returns a slice type converter
// for the given slice or array. Panics if the
// given value is not a slice.
func (v Value) Slice() (s Slice) {
	if k := v.Kind(); k != SLICE && k != ARRAY {
		panic("typ.Value.Slice: not a slice")
	}
	s.Value = v
	return
}

// Map returns a map type converter
// for the given map. Panics if the given
// value is not a map.
func (v Value) Map() (m hmap) {
	if v.Kind() != MAP {
		panic("typ.Value.Map: not a map")
	}
	m.Value = v
	return
}

// ----------------------------------------------------------------------------
// IFACE

// Iface is a type converter for interfaces
type Iface struct {
	Type    *Type
	Pointer unsafe.Pointer
	flag
}

// Reflect returns a reflect.Value from an Iface
func (i Iface) Reflect() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&i))
}

// Value returns a Value type converter
func (i Iface) Value() (v Value) {
	v.Value = i.Reflect()
	return
}

// ----------------------------------------------------------------------------
// SLICE

// slice is a type converter for slices
type Slice struct{ Value }

// Slice returns a slice type converter
// for the given slice or array. Panics if the
// given value is not a slice.
func SliceOf(slice any) (s Slice) {
	s.Value = ValueOf(slice)
	if k := s.Kind(); k != SLICE && k != ARRAY {
		panic("typ: not a slice")
	}
	return
}

// Map converts the slice to a map
func (s Slice) Map() map[string]any {
	hmap := make(map[string]any, s.Len())
	for i := 0; i < s.Len(); i++ {
		hmap[strconv.Itoa(i)] = s.Index(i).Interface()
	}
	return hmap
}

// Scan reads the value from the Slice
// into the given pointer to an Array, Slice, Map, or Struct
func (s Slice) Scan(ptr any) {
	v := ValueOf(ptr)
	if v.Kind() != POINTER {
		panic("typ: not a pointer")
	}
	v = v.Elem()
	len := s.Len()
	if v.Len() < len {
		panic("typ: not enough space")
	}
	switch v.Kind() {
	case ARRAY, SLICE:
	case MAP:
	case STRUCT:
	}
}

// ----------------------------------------------------------------------------
// MAP

// map is a type converter for maps
type hmap struct{ Value }

// Map returns a map type converter
// for the given map. Panics if the given
// value is not a map.
func Map(hmap any) (m hmap) {
	m.Value = ValueOf(hmap)
	if m.Kind() != MAP {
		panic("typ: not a map")
	}
	return
}

// Slice converts the map to a slice
func (m hmap) Slice() []any {
	slice := make([]any, m.Len())
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		slice[i] = iter.Value().Interface()
	}
	return slice
}

// ----------------------------------------------------------------------------
// STRUCT

// struct is a type converter for structs
type Struct struct{ Value }

// Struct returns a struct type converter
// for the given struct. Panics if the given
// value is not a struct.
func StructOf(strct any) (s Struct) {
	s.Value = ValueOf(strct)
	if s.Kind() != STRUCT {
		panic("typ: not a struct")
	}
	return
}

// Slice converts the struct to a slice
func (s Struct) Slice() []any {
	slice := make([]any, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		slice[i] = s.Field(i).Interface()
	}
	return slice
}

// Map converts the struct to a map
func (s Struct) Map(tag ...string) map[string]any {
	hmap := make(map[string]any, s.NumField())
	if len(tag) > 0 {
		tname := tag[0]
		s.Type().ForFields(func(i int, f *FieldType) (brake bool) {
			hmap[f.TagValue(tname)] = s.Field(i).Interface()
			return
		})
		return hmap
	}
	s.Type().ForFields(func(i int, f *FieldType) (brake bool) {
		hmap[f.Name()] = s.Field(i).Interface()
		return
	})
	return hmap
}
