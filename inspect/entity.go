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
	"go/ast"
	"go/token"

	"github.com/jcdotter/go/data"
)

// -----------------------------------------------------------------------------
// ENTITY
// These are the basic building blocks of the
// Go language: package, import, const, var, type, and func.

// Entity keywords
var entityBytes = map[byte][]byte{
	PACKAGE: PackageKey,
	IMPORT:  ImportKey,
	CONST:   ConstKey,
	VAR:     VarKey,
	TYPE:    TypeKey,
	FUNC:    FuncKey,
}

// Entity represents an instance of a Go
// declaration keyword and name.
func Entity(e byte, n string) *entity {
	if _, ok := entityBytes[e]; ok {
		return &entity{Kind: e, Name: []byte(n)}
	}
	panic(ErrInvalidEntity)
}

// Entity represents an instance of a Go
// declaration keyword and name.
type entity struct {
	Kind byte
	Name []byte
}

func (e *entity) String() string {
	return string(entityBytes[e.Kind])
}

// -----------------------------------------------------------------------------
// ENTITIES
// Struct representations of Go declarations.
// These are the basic building blocks of the
// Go language: package, import, const, var, type, and func.

// Package represents a Go package and
// contains the package content, metadata, and
// references to the package files and imported packages.
type Package struct {
	Name    string     // the full package name
	Path    string     // the local directory where the package is located
	Imports *data.Data // the packages imported in the files
	Files   *data.Data // the files in the package
	Values  *data.Data // the declared values in the package
	Types   *data.Data // the declared types in the package
	Funcs   *data.Data // the declared functions in the package
	i       bool       // the package has been inspected
	q       []**Type   // unresolved types found during parsing
}

func NewPackage(path string) *Package {
	return &Package{
		Path:    path,
		Imports: data.Make[*Package](data.Cap),
		Files:   data.Make[*File](data.Cap),
		Values:  data.Make[*Value](data.Cap),
		Types:   data.Make[*Type](data.Cap),
		Funcs:   data.Make[*Func](data.Cap),
	}
}

// data.Elem interface method
func (p *Package) Key() string {
	return p.Path
}

type File struct {
	p *Package   // the parser parsing the file
	n string     // the file name
	i *data.Data // the file imports
	t *ast.File  // the file abstract syntax tree
}

func NewFile(pkg *Package, name string) *File {
	return &File{
		p: pkg,
		n: name,
		i: data.Make[*Import](data.Cap),
	}
}

// data.Elem interface method
func (f *File) Key() string {
	return f.n
}

// Name returns the file name.
func (f *File) Name() string {
	return f.n
}

// Package returns the package object that the file belongs to.
func (f *File) Package() *Package {
	return f.p
}

// Import represents an imported package in a file.
type Import struct {
	file *File    // the file where the import is declared
	name string   // the import alias or pkg suffix
	pkg  *Package // the imported package
}

// data.Elem interface method
func (i *Import) Key() string {
	return i.file.Key() + i.name
}

// Name returns the import name.
func (i *Import) Name() string {
	return i.name
}

// Package returns the imported package.
func (i *Import) Package() *Package {
	return i.pkg
}

// File returns the file where the import is declared.
func (i *Import) File() *File {
	return i.file
}

// Value represents a declared value (const or var) in a file.
type Value struct {
	file *File  // the file where the value is declared
	kind byte   // the value kind (const or var)
	name string // the value name
	typ  *Type  // the value type
}

// data.Elem interface method
func (v *Value) Key() string {
	return v.name
}

// Name returns the value name.
func (v *Value) Name() string {
	return v.name
}

// File returns the file where the value is declared.
func (v *Value) File() *File {
	return v.file
}

// Type returns the value type.
func (v *Value) Type() *Type {
	return v.typ
}

// Kind returns the value kind (CONST or VAR).
func (v *Value) Kind() byte {
	return v.kind
}

// Type represents a declared type in a file.
type Type struct {
	file   *File   // he file where the type is declared
	name   string  // the type name
	imp    *Import // the type source if imported
	kind   byte    // the type kind
	object Object  // the type object, if an object type
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

// data.Elem interface method
func (t *Type) Key() string {
	return t.name
}

// Name returns the type name.
func (t *Type) Name() string {
	return t.name
}

// File returns the file where the type is declared.
func (t *Type) File() *File {
	return t.file
}

// Import returns the import source of the type, if the type is imported.
func (t *Type) Import() *Import {
	return t.imp
}

// Kind returns the type kind.
func (t *Type) Kind() byte {
	return t.kind
}

// Object returns the type object, if the type is an object type.
func (t *Type) Object() Object {
	return t.object
}

// Func represents a declared function in a file.
type Func struct {
	file    *File      // the file where the function is declared
	comment string     // the function comment
	name    string     // the function name
	typ     *Type      // the function type
	of      *Type      // the function receiver type
	in      *data.Data // the function input parameters
	out     *data.Data // the function output parameters
}

// data.Elem interface method
func (f *Func) Key() string {
	return f.name
}
