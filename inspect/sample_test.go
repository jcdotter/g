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
	_ "testing"
)

const (
	Const0 = iota
	Const1
	Const2 byte = iota
	Const3
)

// basiclit types
var (
	IntBasic     = 1
	FloatBasic   = 1.0
	ComplexBasic = 1.0i
	StringBasic  = "string"
	RuneBasic    = 'r'
)

// paren types
var (
	IntParen     = (1)
	FloatParen   = (1.0)
	ComplexParen = (1.0i)
	StringParen  = ("string")
	RuneParen    = ('r')
)

// pointer types
var (
	IntPointer     *int
	FloatPointer   *float64
	ComplexPointer *complex128
	StringPointer  *string
	RunePointer    *rune

	IntRef     = &IntBasic
	FloatRef   = &FloatBasic
	ComplexRef = &ComplexBasic
	StringRef  = &StringBasic
	RuneRef    = &RuneBasic
)

// binary types
var (
	IntBinary    = 1 << 1
	FloatBinary  = 1.0 + 1.0
	StringBinary = "string" + "string"
)

// call types
var (
	IntCall  = FuncLit()
	ByteCall = byte(IntCall)
)

// funclit types
var (
	FuncLit = func() int { return 1 }
)

/* type Str string

type S struct {
	X string
}

type T struct {
	X string
	Y S
	Z []Any
}

func (t T) C() any {
	return t.Y
}

func (s S) D() any {
	return s.X
}

var A = S{X: "string"}
var B = T{X: "string", Y: A}
var C = func() T { return B }

func D() T { return B }

var X = A.X


type Any interface{ Key() string } */

//var M = strings.Index("string", "s")

/*
var W = B.C()
var X = (&B).X
var Y = (D()).X
var Z = C().Y.X
var V = B.C().(S).X

var O = (func() S { return S{X: "string"} })().X

var M *data.Data */

/*
var NoType = 1
var Values1, Values2 int = 1, 2
var Value3 error = nil
var Value4 = "string"
var Value5 = &Values2
var Value6 = (int32)(1)

// var Value7, Value8 = (Func1()) // test when inspect func is implemented
var Value9 = 1 << 1
var Value10 = true
var Value11 = ("string") + "string"
var Value12 = true == false
var Value13 = [3]string{"one", "two", "three"}
var Value14 = func(one, two string) (three int, four int) { return 1, 2 }

const (
	Const1 uint8 = iota
	Const2
)

var (
	Var1,
	Var2 string
)

func Func1() (string, string) {
	return "1", "string"
}

func Func2() int {
	return 1
} */

/*
var (
	Bool             bool       = true
	Int              int        = 1
	Uint             uint       = 1
	Float            float64    = 1.0
	Complex          complex128 = 1.0
	String           string     = "string"

	BoolParen    (bool)       = (bool)(true)
	IntParen     (int)        = (int)(1)
	UintParen    (uint)       = (uint)(1)
	FloatParen   (float64)    = (float64)(1.0)
	ComplexParen (complex128) = (complex128)(1.0)
	StringParen  (string)     = (string)("string")

	BoolParenVal   = (true)
	IntParenVal    = (1)
	FloatParenVal  = (1.0)
	StringParenVal = ("string")

	VarArray [3]int         = [3]int{1, 2, 3}
	VarSlice []int          = []int{1, 2, 3}
	VarMap   map[string]int = map[string]int{"one": 1, "two": 2, "three": 3}
	VarChan  chan string    = make(chan string)
	VarFunc  func(s string) (i int, e error)

	Ref *int = &Int

	Number = 1 >=
		2+ // comment
			3
)

var TestVar = "TEST"

type TEST struct {
	Bool bool
}

type (
	BoolType    bool
	IntType     int
	UintType    uint
	FloatType   float64
	ComplexType complex128
	StringType  string
	FuncType    func(StringType) (IntType, error)
	StructType  struct {
		BoolType
		IntType
		StringType
	}
	StructType2   struct{ Bool, Int, String StringType }
	ArrayType     [3]IntType
	SliceType     []IntType
	MapType       map[StringType]IntType
	ChanType      chan StringType
	InterfaceType interface {
		Bool() bool
		Int() int
		String() StringType
	}
	PointerType **IntType
)

func FuncTest() string {
	return "TEST"
}
*/
