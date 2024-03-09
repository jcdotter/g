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

// general expression types
var (
	IntParen     = (1)                                                       // parenthesized expression
	IntPointer   *int                                                        // pointer (star) expression
	IntRef       = &IntBasic                                                 // reference (unary) expression
	StringBinary = "string" + "string"                                       // binary expression
	IntCall      = FuncLit()                                                 // call expression
	FuncLit      = func() int { return 1 }                                   // func literal
	ArrayLit     = [3]int{1, 2, 3}                                           // array literal
	SliceLit     = []int{1, 2, 3}                                            // slice literal
	MapLit       = map[string]int{"one": 1, "two": 2, "three": 3}            // map literal
	ChanLit      chan string                                                 // chan literal
	StructLit    = struct{ Bool, Int, String string }{"true", "1", "string"} // struct literal
	SelExpr      = StructLit.Int                                             // selector expression
	IndexExpr    = ArrayLit[1]                                               // index expression
)

// composite types
type (
	BoolType                    bool
	InterfaceType               interface{ Bool() bool }
	StructType[T InterfaceType] struct {
		Bool   BoolType
		Int    int
		String string
		Iface  T
		Func   func(string) (int, error)
	}
)

// functions
func (b *BoolType) BoolMethod(i int) int        { return i }
func (s *StructType[T]) StructMethod(i int) int { return i }
