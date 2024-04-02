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
	"reflect"
	"strconv"
)

// ----------------------------------------------------------------------------
// SLICE

// slice is a type converter for slices
type slice struct {
	reflect.Value
}

// Slice returns a slice type converter
// for the given slice or array. Panics if the
// given value is not a slice.
func Slice(slice any) (s slice) {
	s.Value = reflect.ValueOf(slice)
	if k := s.Kind(); k != reflect.Slice && k != reflect.Array {
		panic("encoder: not a slice")
	}
	return
}

// Map converts the slice to a map
func (s slice) Map() map[string]any {
	hmap := make(map[string]any, s.Len())
	for i := 0; i < s.Len(); i++ {
		hmap[strconv.Itoa(i)] = s.Index(i).Interface()
	}
	return hmap
}

// ----------------------------------------------------------------------------
// MAP

// map is a type converter for maps
type hmap struct {
	reflect.Value
}

// Map returns a map type converter
// for the given map. Panics if the given
// value is not a map.
func Map(hmap any) (m hmap) {
	m.Value = reflect.ValueOf(hmap)
	if m.Kind() != reflect.Map {
		panic("encoder: not a map")
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
