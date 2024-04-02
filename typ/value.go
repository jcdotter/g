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
	"unsafe"
)

// ------------------------------------------------------------ /
// Value IMPLEMENTATION
// inspired by golang standard reflect.Value
// with expanded methods and type conversations

// Value contains a pointer to a value and to its datatype
type Value struct {
	typ *Type
	ptr unsafe.Pointer
	flag
}

func ValueOf(a any) Value {
	v := reflect.ValueOf(a)
	return FromReflect(v)
}

func ValueOfV(a any) Value {
	if n, is := a.(Value); !is {
		return ValueOf(a)
	} else {
		return n
	}
}

func FromReflect(v reflect.Value) Value {
	return *(*Value)(unsafe.Pointer(&v))
}

func (v Value) Reflect() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&v))
}

func (v Value) Init() Value {
	if v.ptr == nil {
		return v.typ.NewValue().Elem()
	}
	k := v.Kind()
	if (k == MAP || k == SLICE || k == POINTER) && *(*unsafe.Pointer)(v.ptr) == nil {
		*(*unsafe.Pointer)(v.ptr) = *(*unsafe.Pointer)(v.typ.NewValue().Elem().ptr)
	}
	return v
}
