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

type Object interface{}
type Pointer struct {
	typ  *Type
	elem *Type
}
type Array struct {
	typ  *Type
	elem *Type
	len  int
}
type Map struct{}
type Chan struct{}
type Struct struct{}
type Field struct{}
