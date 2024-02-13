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

// go datatype kinds
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

// go datatype aliases
const (
	BYTE  byte = 30 + iota // alias for UINT8
	RUNE                   // alias for INT32
	ANY                    // alias for INTERFACE
	ERROR                  // alias for interface{Error() string}
	ELIPS                  // alias for ...
	TYPE                   // alias for Type
	BYTES                  // alias for []byte
	TIME                   // alias for struct{uint64, int64, uintptr}
	UUID                   // alias for [16]byte
)

var kindNames = []string{
	INVALID:       "invalid",
	BOOL:          "bool",
	INT:           "int",
	INT8:          "int8",
	INT16:         "int16",
	INT32:         "int32",
	INT64:         "int64",
	UINT:          "uint",
	UINT8:         "uint8",
	UINT16:        "uint16",
	UINT32:        "uint32",
	UINT64:        "uint64",
	UINTPTR:       "uintptr",
	FLOAT32:       "float32",
	FLOAT64:       "float64",
	COMPLEX64:     "complex64",
	COMPLEX128:    "complex128",
	ARRAY:         "array",
	CHAN:          "chan",
	FUNC:          "func",
	INTERFACE:     "interface",
	MAP:           "map",
	POINTER:       "pointer",
	SLICE:         "slice",
	STRING:        "string",
	STRUCT:        "struct",
	UNSAFEPOINTER: "unsafe.Pointer",
	BYTE:          "byte",
	RUNE:          "rune",
	ANY:           "any",
	ERROR:         "error",
	ELIPS:         "...",
	TYPE:          "type",
	BYTES:         "[]byte",
	TIME:          "time",
	UUID:          "uuid",
}

const (
	kindMask             = (1 << 5) - 1
	KindDirectIface      = 1 << 5
	flagStickyRO    flag = 1 << 5
	flagEmbedRO     flag = 1 << 6
	flagIndir       flag = 1 << 7
	flagAddr        flag = 1 << 8

	tflagUncommon tflag = 1 << 0
)

type flag uintptr
type tflag uint8
