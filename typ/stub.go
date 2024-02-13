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

//go:noescape
//go:linkname toType reflect.toType
func toType(t *Type) reflect.Type

//go:noescape
//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*Type) unsafe.Pointer

//go:noescape
//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(*Type, int) unsafe.Pointer

//go:noescape
//go:linkname makemap runtime.makemap
func makemap(t *Type, cap int, hint unsafe.Pointer) unsafe.Pointer

//go:noescape
//go:linkname resolveNameOff reflect.resolveNameOff
func resolveNameOff(ptrInModule unsafe.Pointer, off int32) unsafe.Pointer
