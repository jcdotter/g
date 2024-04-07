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
	return (*(*Iface)(unsafe.Pointer(&v.Value))).Type
}

// Pointer returns the pointer to the value
func (v Value) Pointer() unsafe.Pointer {
	return (*(*Iface)(unsafe.Pointer(&v.Value))).Pointer
}

// Uintptr returns the value as a uintptr
func (v Value) Uintptr() uintptr {
	return uintptr(v.Pointer())
}

// Elem returns the value that the pointer points to
func (v Value) Elem() Value {
	return FromReflectValue(v.Value.Elem())
}

// Kind returns the kind of the value
func (v Value) Kind() byte {
	return byte(v.Value.Kind())
}

// KindX returns the extended kind of the value
// to include aliases and special types
func (v Value) KindX() byte {
	return v.Type().KindX()
}

// SetType sets the actual data type of interface Value
func (v Value) SetType() Value {
	if v.Kind() == INTERFACE {
		var e *Value
		if (*interfaceType)(unsafe.Pointer(v.Type())).NumMethod() != 0 {
			e = (*Value)(unsafe.Pointer((*interface{ M() })(v.Pointer())))
		} else {
			e = (*Value)(v.Pointer())
		}
		if e.Type() != nil && e.Type().kind != 0 {
			i := Iface{e.Type(), e.Pointer(), e.Type().flag()}
			return Value{*(*reflect.Value)(unsafe.Pointer(&i))}
		}
	}
	return v
}

// IsNil returns true if the value is nil
func (v Value) IsNil() bool {
	switch v.Kind() {
	case CHAN, FUNC, MAP, POINTER, SLICE, INTERFACE, UNSAFEPOINTER:
		return v.Value.IsNil()
	}
	return false
}

// Len returns the length of the value
func (v Value) Len() int {
	if v.Kind() == STRUCT {
		return v.Type().NumField()
	}
	return v.Value.Len()
}

// Slice returns a slice type converter
// for the given slice or array. Panics if the
// given value is not a slice.
func (v Value) Slice() (s Slice) {
	switch v.Kind() {
	case SLICE, ARRAY:
		s.Value = v
	case POINTER:
		s.Value = v.Elem()
	default:
		panic("typ.Value.Slice: not a slice")
	}
	return
}

// Map returns a map type converter
// for the given map. Panics if the given
// value is not a map.
func (v Value) Map() (m Map) {
	switch v.Kind() {
	case MAP:
		m.Value = v
	case POINTER:
		m.Value = v.Elem()
	default:
		panic("typ.Value.Map: not a map")
	}
	return
}

// Struct returns a struct type converter
// for the given struct. Panics if the given
// value is not a struct.
func (v Value) Struct() (s Struct) {
	switch v.Kind() {
	case STRUCT:
		s.Value = v
	case POINTER:
		s.Value = v.Elem()
	default:
		panic("typ.Value.Struct: not a struct")
	}
	return
}

// Binary returns a binary type converter
func (v Value) Binary() Binary {
	switch v.Kind() {
	case STRING:
		return Binary(v.String())
	case SLICE:
		switch v.Type().Elem().Kind() {
		case UINT8:
			return Binary(string(v.Value.Bytes()))
		case INT32:
			return *(*Binary)(v.Pointer())
		}
	}
	panic("typ.Value.Binary: not a string or byte slice")
}

// Bytes returns the value as a byte slice
func (v Value) Bytes() []byte {
	switch v.Kind() {
	case STRING:
		return []byte(v.String())
	case SLICE:
		if v.Type().Elem().Kind() == UINT8 {
			return v.Value.Bytes()
		}
	}
	panic("typ.Value.Bytes: not a string or byte slice")
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
func SliceOf(slice any) Slice {
	return ValueOf(slice).Slice()
}

// Extend increases the length of the slice
// by n elements. Panics if the given value
// is not a addressable slice.
func (s Slice) Extend(n int) Slice {
	s.Grow(n)
	*(*int)(unsafe.Pointer(uintptr(s.Pointer()) + 8)) = s.Cap()
	return s
}

// ForEach iterates over the slice and calls
// the given function for each element.
func (s Slice) ForEach(fn func(i int, v Value) (brake bool)) {
	for i := 0; i < s.Len(); i++ {
		if fn(i, FromReflectValue(s.Index(i))) {
			break
		}
	}
}

// Slice converts the slice to a slice
func (s Slice) Slice() []any {
	slice := make([]any, s.Len())
	for i := 0; i < s.Len(); i++ {
		slice[i] = s.Index(i).Interface()
	}
	return slice
}

// Map converts the slice to a map
func (s Slice) Map() map[string]any {
	hmap := make(map[string]any, s.Len())
	for i := 0; i < s.Len(); i++ {
		hmap[strconv.Itoa(i)] = s.Index(i).Interface()
	}
	return hmap
}

// Scan reads the value from the Slice into the given
// dest pointer to an Array, Slice, Map, or Struct
func (s Slice) Scan(dest any) {
	v := prepScanDest(s.Value, dest)
	switch v.Kind() {
	case ARRAY, SLICE:
		d := v.Slice()
		s.ForEach(func(i int, e Value) (brake bool) {
			d.Index(i).Set(e.Reflect())
			return
		})
	case MAP:
		d := v.Map()
		s.ForEach(func(i int, e Value) (brake bool) {
			d.SetMapIndex(reflect.ValueOf(i), e.Reflect())
			return
		})
	case STRUCT:
		d := v.Struct()
		s.ForEach(func(i int, e Value) (brake bool) {
			d.Field(i).Set(e.Reflect())
			return
		})
	}
}

// ----------------------------------------------------------------------------
// MAP

// map is a type converter for maps
type Map struct{ Value }

// Map returns a map type converter
// for the given map. Panics if the given
// value is not a map.
func MapOf(hmap any) (m Map) {
	return ValueOf(hmap).Map()
}

// ForEach iterates over the map and calls
// the given function for each element.
func (m Map) ForEach(fn func(k, v Value) (brake bool)) {
	iter := m.MapRange()
	for iter.Next() {
		if fn(FromReflectValue(iter.Key()), FromReflectValue(iter.Value())) {
			break
		}
	}
}

// Keys returns the keys of the map
func (m Map) KeyVals() (k []string, v []any) {
	k = make([]string, m.Len())
	v = make([]any, m.Len())
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		k[i] = iter.Key().String()
		v[i] = iter.Value().Interface()
	}
	return
}

// Keys returns the keys of the map
func (m Map) Keys() []string {
	keys := make([]string, m.Len())
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		keys[i] = iter.Key().String()
	}
	return keys
}

// Slice converts the map to a slice
func (m Map) Values() []any {
	slice := make([]any, m.Len())
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		slice[i] = iter.Value().Interface()
	}
	return slice
}

// Map converts the map to a map
func (m Map) Map() map[string]any {
	hmap := make(map[string]any, m.Len())
	m.ForEach(func(k, v Value) (brake bool) {
		hmap[k.String()] = v.Interface()
		return
	})
	return hmap
}

// Scan reads the value from the Map into the given
// dest pointer to an Array, Slice, Map, or Struct
func (m Map) Scan(dest any, tag ...string) {
	v := prepScanDest(m.Value, dest)
	switch v.Kind() {
	case ARRAY, SLICE:
		d, i := v.Slice(), 0
		m := m.MapRange()
		for m.Next() {
			d.Index(i).Set(m.Value())
			i++
		}
	case MAP:
		d := v.Map()
		m.ForEach(func(k, v Value) (brake bool) {
			d.SetMapIndex(k.Reflect(), v.Reflect())
			return
		})
	case STRUCT:
		d := v.Struct()
		if len(tag) > 0 {
			if tags, has := d.Type().TagValues(tag[0]); has {
				m.ForEach(func(k, v Value) (brake bool) {
					for i, tag := range tags {
						if k.String() == tag {
							d.Field(i).Set(v.Reflect())
							break
						}
					}
					return
				})
				return
			}
		}
		m.ForEach(func(k, v Value) (brake bool) {
			d.FieldByName(k.String()).Set(v.Reflect())
			return
		})
	}
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

// ForEach iterates over the struct and calls
// the given function for each field.
func (s Struct) ForEach(fn func(i int, f *FieldType, v Value) (brake bool)) {
	s.Type().ForFields(func(i int, f *FieldType) (brake bool) {
		return fn(i, f, FromReflectValue(s.Field(i)))
	})
}

// FieldByTag returns the field by the given tag value
func (s Struct) FieldByTag(tag, val string) Value {
	var field Value
	s.Type().ForFields(func(i int, f *FieldType) (brake bool) {
		if f.TagValue(tag) == val {
			field = FromReflectValue(s.Field(i))
			return true
		}
		return
	})
	return field
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

// Scan reads the value from the Struct into the given
// dest pointer to an Array, Slice, Map, or Struct
func (s Struct) Scan(dest any, tag ...string) {
	v := prepScanDest(s.Value, dest)
	switch v.Kind() {
	case ARRAY, SLICE:
		d := v.Slice()
		s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
			d.Index(i).Set(v.Reflect())
			return
		})
	case MAP:
		d := v.Map()
		// Use tag values to map fields
		if len(tag) > 0 {
			if tags, has := d.Type().TagValues(tag[0]); has {
				s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
					d.SetMapIndex(reflect.ValueOf(tags[i]), v.Reflect())
					return
				})
				return
			}
		}
		// Use field names to map fields
		s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
			d.SetMapIndex(reflect.ValueOf(f.Name()), v.Reflect())
			return
		})
	case STRUCT:
		d := v.Struct()
		if len(tag) > 0 {
			// Use tag values to map fields
			if tags, has := d.Type().TagValues(tag[0]); has {
				s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
					for i, tag := range tags {
						if f.TagValue(tag) == tag {
							d.Field(i).Set(v.Reflect())
							break
						}
					}
					return
				})
				return
				// Use field names to map fields
			} else if tag[0] == "Name" {
				s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
					d.FieldByName(f.Name()).Set(v.Reflect())
					return
				})
				return
			}
		}
		// map fields by index
		s.ForEach(func(i int, f *FieldType, v Value) (brake bool) {
			d.Field(i).Set(v.Reflect())
			return
		})
	}
}

// ----------------------------------------------------------------------------
// BINARY

type Binary []rune

// Binary returns a binary type converter
func BinaryOf(binary any) Binary {
	switch b := binary.(type) {
	case Binary:
		return b
	case []rune:
		return b
	case []byte:
		return Binary(string(b))
	case string:
		return Binary(b)
	default:
		panic("typ.BinaryOf: not a binary")
	}
}

// Bytes returns the binary as a byte slice
func (b Binary) Bytes() []byte {
	return []byte(string(b))
}

// String returns the binary as a string
func (b Binary) String() string {
	return string(b)
}
