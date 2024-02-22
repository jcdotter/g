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

import "github.com/jcdotter/go/data"

type Object interface{}

type Array struct {
	typ  *Type
	elem *Type
	len  int
}

type Struct struct {
	typ     *Type
	fields  *data.Data
	methods *data.Data
}

func NewStruct(typ *Type) *Struct {
	return &Struct{
		typ:     typ,
		fields:  data.Make[*Field](data.Cap),
		methods: data.Make[*Func](data.Cap),
	}
}

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

type Pointer struct{ typ, elem *Type }
type Map struct{ typ, key, elem *Type }
type Chan struct{}
