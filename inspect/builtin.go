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
	"go/token"
	"os"

	"github.com/jcdotter/go/data"
)

var (
	PkgPath = os.Getenv("HOME") + "/GO/pkg/mod/"
	SrcPath = "/usr/local/go/src/"
)

// -----------------------------------------------------------------------------
// KINDS

// Kind is an enumerated representaion of the basic builtin datatypes
const (
	INVALID byte = iota
	BOOL
	INT
	INT8
	INT16
	INT32
	INT64
	UINT
	UINT8
	UINT16
	UINT32
	UINT64
	UINTPTR
	FLOAT32
	FLOAT64
	COMPLEX64
	COMPLEX128
	ARRAY
	CHAN
	FUNC
	INTERFACE
	MAP
	POINTER
	SLICE
	STRING
	STRUCT
	UNSAFEPOINTER
)

// kind aliases
const (
	BYTE byte = iota + 30
	RUNE
	ANY
	ERROR
	ELIPS
	TYPE
	ORDERED
	INTEGER
	FLOAT
	COMPLEX
)

// entity declaration kind
const (
	PACKAGE byte = iota + 50
	IMPORT
	CONST
	VAR
	// TYPE = 35 above
	// FUNC = 18 above
)

// -----------------------------------------------------------------------------
// BUILTINS
// implements the standard builtin types and functions
// that don't require an import

var BuiltinPkg = &Package{
	Name:   "builtin",
	Path:   PkgPath,
	Values: BuiltinValues,
	Types:  BuiltinTypes,
	Funcs:  BuiltinFuncs,
}

var BuiltinFile = &File{
	n: "builtin.go",
}

var BuiltinImport = &Import{
	name: "builtin",
	pkg:  BuiltinPkg,
}

var BuiltinValues = data.Of(
	&Value{file: BuiltinFile, name: "nil", kind: INVALID},
	&Value{file: BuiltinFile, name: "true", kind: BOOL},
	&Value{file: BuiltinFile, name: "false", kind: BOOL},
	&Value{file: BuiltinFile, name: "iota", kind: INT},
)

var BuiltinTypes = data.Of(
	&Type{file: BuiltinFile, name: "nil", kind: INVALID},
	&Type{file: BuiltinFile, name: "bool", kind: BOOL},
	&Type{file: BuiltinFile, name: "int", kind: INT},
	&Type{file: BuiltinFile, name: "int8", kind: INT8},
	&Type{file: BuiltinFile, name: "int16", kind: INT16},
	&Type{file: BuiltinFile, name: "int32", kind: INT32},
	&Type{file: BuiltinFile, name: "int64", kind: INT64},
	&Type{file: BuiltinFile, name: "uint", kind: UINT},
	&Type{file: BuiltinFile, name: "uint8", kind: UINT8},
	&Type{file: BuiltinFile, name: "uint16", kind: UINT16},
	&Type{file: BuiltinFile, name: "uint32", kind: UINT32},
	&Type{file: BuiltinFile, name: "uint64", kind: UINT64},
	&Type{file: BuiltinFile, name: "uintptr", kind: UINTPTR},
	&Type{file: BuiltinFile, name: "float32", kind: FLOAT32},
	&Type{file: BuiltinFile, name: "float64", kind: FLOAT64},
	&Type{file: BuiltinFile, name: "complex64", kind: COMPLEX64},
	&Type{file: BuiltinFile, name: "complex128", kind: COMPLEX128},
	nil, //&Type{file: BuiltinFile, name: "array", kind: ARRAY},
	nil, //&Type{file: BuiltinFile, name: "chan", kind: CHAN},
	nil, //&Type{file: BuiltinFile, name: "func", kind: FUNC},
	nil, //&Type{file: BuiltinFile, name: "interface", kind: INTERFACE},
	nil, //&Type{file: BuiltinFile, name: "map", kind: MAP},
	nil, //&Type{file: BuiltinFile, name: "pointer", kind: POINTER},
	nil, //&Type{file: BuiltinFile, name: "slice", kind: SLICE},
	&Type{file: BuiltinFile, name: "string", kind: STRING},
	nil, //&Type{file: BuiltinFile, name: "struct", kind: STRUCT},
	&Type{file: BuiltinFile, name: "unsafe.Pointer", kind: UNSAFEPOINTER},
	nil,
	nil,
	nil,
	// aliases
	&Type{file: BuiltinFile, name: "byte", kind: BYTE},
	&Type{file: BuiltinFile, name: "rune", kind: RUNE},
	&Type{file: BuiltinFile, name: "any", kind: ANY},
	&Type{file: BuiltinFile, name: "error", kind: ERROR},
	nil, //&Type{file: BuiltinFile, name: "elips", kind: ELIPS},
)

// custom builtin types
var (
	_Type         = &Type{file: BuiltinFile, name: "type", kind: TYPE}
	_TypePtr      = &Type{file: BuiltinFile, name: "*type", kind: POINTER}
	_TypeMap      = &Type{file: BuiltinFile, name: "map[type]type", kind: MAP}
	_TypeSlice    = &Type{file: BuiltinFile, name: "[]type", kind: SLICE}
	_TypeElips    = &Type{file: BuiltinFile, name: "...type", kind: ELIPS}
	_TypeChan     = &Type{file: BuiltinFile, name: "chan type", kind: CHAN}
	_TypeMapSlice = &Type{file: BuiltinFile, name: "[]type | map[type]type", kind: INTERFACE}
	_Ordered      = &Type{file: BuiltinFile, name: "ordered", kind: ORDERED}
	_OrderedElips = &Type{file: BuiltinFile, name: "...ordered", kind: ELIPS}
	_IntElips     = &Type{file: BuiltinFile, name: "...integer", kind: ELIPS}
	_Float        = &Type{file: BuiltinFile, name: "float", kind: FLOAT}
	_Complex      = &Type{file: BuiltinFile, name: "complex", kind: COMPLEX}
)

var BuiltinFuncs = data.Of(
	&Func{file: BuiltinFile, name: "append", in: data.Of(_TypeSlice, _TypeElips), out: data.Of(_TypeSlice)},
	&Func{file: BuiltinFile, name: "cap", in: data.Of(_Type), out: data.Of(BuiltinTypes.List()[INT])},
	&Func{file: BuiltinFile, name: "clear", in: data.Of(_TypeMapSlice)},
	&Func{file: BuiltinFile, name: "close", in: data.Of(_TypeChan)},
	&Func{file: BuiltinFile, name: "complex", in: data.Of(_Float, _Float), out: data.Of(_Complex)},
	&Func{file: BuiltinFile, name: "copy", in: data.Of(_TypeSlice, _TypeSlice), out: data.Of(BuiltinTypes.List()[INT])},
	&Func{file: BuiltinFile, name: "delete", in: data.Of(_TypeMap, _Type)},
	&Func{file: BuiltinFile, name: "imag", in: data.Of(_Complex), out: data.Of(_Float)},
	&Func{file: BuiltinFile, name: "len", in: data.Of(_Type), out: data.Of(BuiltinTypes.List()[INT])},
	&Func{file: BuiltinFile, name: "make", in: data.Of(_Type, _IntElips), out: data.Of(_Type)},
	&Func{file: BuiltinFile, name: "max", in: data.Of(_Ordered, _OrderedElips), out: data.Of(_Ordered)},
	&Func{file: BuiltinFile, name: "min", in: data.Of(_Ordered, _OrderedElips), out: data.Of(_Ordered)},
	&Func{file: BuiltinFile, name: "new", in: data.Of(_Type), out: data.Of(_TypePtr)},
	&Func{file: BuiltinFile, name: "panic", in: data.Of(BuiltinTypes.List()[ANY])},
	&Func{file: BuiltinFile, name: "print", in: data.Of(_TypeElips)},
	&Func{file: BuiltinFile, name: "println", in: data.Of(_TypeElips)},
	&Func{file: BuiltinFile, name: "real", in: data.Of(_Complex), out: data.Of(_Float)},
	&Func{file: BuiltinFile, name: "recover", out: data.Of(BuiltinTypes.List()[ANY])},
)

func init() {
	// set paths if available
	if p := os.Getenv("GOPATH"); p != "" {
		PkgPath = p + "/pkg/mod/"
	}
	if p := os.Getenv("GOROOT"); p != "" {
		SrcPath = p + "/src/"
	}
	// connect builtins
	BuiltinPkg.Files = data.Of(BuiltinFile)
	BuiltinFile.p = BuiltinPkg
	// add types to buildin values
	types := BuiltinTypes.List()
	for _, v := range BuiltinValues.List() {
		v.(*Value).typ = types[v.(*Value).kind].(*Type)
	}
}

func TypeToken(t token.Token) *Type {
	switch t {
	case token.INT:
		return BuiltinTypes.Get("int").(*Type)
	case token.FLOAT:
		return BuiltinTypes.Get("float64").(*Type)
	case token.IMAG:
		return BuiltinTypes.Get("complex128").(*Type)
	case token.STRING:
		return BuiltinTypes.Get("string").(*Type)
	case token.CHAR:
		return BuiltinTypes.Get("rune").(*Type)
	}
	return nil
}
