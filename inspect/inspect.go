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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"

	"github.com/jcdotter/go/path"
)

// Inspect parses the package content in the path provided and
// returns the package object for inspection, or an error if
// the package cannot be parsed.
func Inspect(PkgPath string) (*Package, error) {
	p := NewPackage(PkgPath)
	if err := p.Parse(); err != nil {
		return nil, err
	}
	for _, f := range p.Files.List() {
		if err := f.(*File).Inspect(); err != nil {
			return nil, err
		}
	}
	p.i = true
	return p, nil
}

// Parse parses the package content if not already parsed. If Entites are provided,
// the package will only parse the provided entities, otherwise the package will
// parse all entities in the package. Returns an error if the package cannot be parsed.
// TODO: Make file parsing concurrent.
func (p *Package) Parse() (err error) {
	// parse each file in the package
	for _, f := range path.Files(p.Path) {
		var file *File

		// parse file name
		n := f[strings.LastIndex(f, "/")+1 : strings.LastIndex(f, ".")]

		// check if file is already parsed
		// else add a new file to the package
		if f := p.Files.Get(n); f != nil {
			return
		}
		file = NewFile(p, n)
		p.Files.Add(file)

		// parse file to abstract syntax tree
		file.t, err = parser.ParseFile(token.NewFileSet(), f, nil, parser.SkipObjectResolution)
		if err != nil {
			return
		}
	}

	// parse package name
	if p.Files.Len() > 0 {
		p.Name = p.Files.Index(0).(*File).t.Name.Name
	}
	return
}

// Inspect inspects the declared entities in the file and
// adds them to the package.
func (f *File) Inspect() (err error) {
	if f.t != nil {
		for _, d := range f.t.Decls {

			// route declaration to appropriate
			// inspection method
			switch decl := d.(type) {
			case *ast.FuncDecl:
				err = f.InspectFunc(decl)
			case *ast.GenDecl:
				switch decl.Tok {
				case token.CONST:
					err = f.InspectValues(CONST, decl.Specs)
				case token.VAR:
					err = f.InspectValues(VAR, decl.Specs)
				case token.TYPE:
					err = f.InspectType(decl.Specs)
				case token.IMPORT:
					err = f.InspectImports(decl.Specs)
				}
			}
		}
	}
	return
}

// InspectImports inspects the import declarations in the file and
// adds them to the package.
func (f *File) InspectImports(specs []ast.Spec) (err error) {
	for _, s := range specs {

		// create and add import to file
		i := s.(*ast.ImportSpec)
		imp := &Import{file: f, name: i.Name.Name}
		f.i.Add(imp)

		// get package by path if already imported another file,
		// otherwise create a new imported package and add it to
		// the current package
		pkgPath := i.Path.Value
		if pkg := f.p.Imports.Get(pkgPath); pkg != nil {
			imp.pkg = pkg.(*Package)
		} else {
			imp.pkg = NewPackage(pkgPath)
			f.p.Imports.Add(imp.pkg)
		}
	}
	return
}

// InspectValues inspects the value declarations in the file and
// adds them to the package.
func (f *File) InspectValues(k byte, specs []ast.Spec) (err error) {

	// the prior type is used when only one type
	// is used for a var or const declaration.
	var priorType *Type

	// iterate through specs and create values for each
	for _, s := range specs {

		// assert value spec
		var vals = s.(*ast.ValueSpec)
		var num = len(vals.Names)
		var names = make([]*ast.Ident, num)

		// check if the value has already been
		// added to the package, if so, skip it
		for i, n := range vals.Names {
			if f.p.Values.Get(n.Name) == nil {
				names[i] = n
				num--
			}
		}
		if num == 0 {
			continue
		}

		// build the value and add it to the package
		var Type *Type

		// get declared type if it exists. if not,
		// the prior declared type will set the
		// type of this value if there are no
		// assginment expressions in this value.
		if vals.Type != nil {
			Type = f.TypeExpr(vals.Type)
			priorType = Type
		} else if priorType != nil && len(vals.Values) == 0 {
			Type = priorType
		}

		// iterate through and create value for each named
		// value in the value spec, using the declared type
		// if it exists, or deriving the type from the
		// value expression if it exists.
		for i, n := range names {

			// if value already exists, name will be nil
			// so skip it and continue to the next value
			if n == nil {
				continue
			}

			// create and add value to package
			val := &Value{file: f, kind: k, name: n.Name}
			f.p.Values.Add(val)

			// set value type if already declared
			// or derive it from value expression
			if Type != nil {
				val.typ = Type
			} else if len(vals.Values) > i {
				val.typ = f.TypeExpr(vals.Values[i])
			}

			// print test
			f.PrintValue(val)
		}

	}
	return
}

func (f *File) InspectValue(k byte, v *ast.ValueSpec) (err error) {
	return f.InspectValues(k, []ast.Spec{v})
}

func (f *File) PrintValue(v *Value) {
	var tname string
	var tkind byte
	if v.typ != nil {
		tname = v.typ.name
		tkind = v.typ.kind
	}
	fmt.Println("VALUE:",
		v.kind,
		v.name,
		tname,
		tkind,
	)
}

func (f *File) InspectType(t []ast.Spec) (err error) {
	/* for _, s := range t {
		fmt.Println("TYPE:", s.(*ast.TypeSpec).Name.Name)
	} */
	return
}

func (f *File) InspectFunc(fn *ast.FuncDecl) (err error) {
	//fmt.Println("FUNC DECL:", fn.Name.Name)
	return
}

func (f *File) GetType(name string) (typ *Type, err error) {

	// check builtin types
	if t := BuiltinTypes.Get(name); t != nil {
		return t.(*Type), nil
	}

	// check declared types
	// TODO: if declared type has not been parsed,
	// follow the ident object to import the
	// declared type in the current package
	if t := f.p.Types.Get(name); t != nil {
		return t.(*Type), nil
	}

	// if an imported type, get the type from the
	// imported package and return it. if the imported
	// package has not been parsed, parse it and return
	// the type if it exists.
	if parts := strings.Split(name, "."); len(parts) > 1 {
		// get imported package if type contains a period
		if imp := f.i.Get(parts[0]); imp != nil {
			i := imp.(*Import)
			// get type from imported package
			t := i.pkg.Types.Get(parts[1])
			if t == nil {
				if err = i.pkg.Parse(); err != nil {
					return
				}
				t = i.pkg.Types.Get(parts[1])
			}
			if t != nil {
				return t.(*Type), nil
			}
		}
		return nil, ErrNotType
	}

	// TODO: parse complex data types by
	// cascading through the type Ident and
	// fetching or parsing the type component
	// and add them to the package types

	return nil, nil
}

func (f *File) TypeExpr(e ast.Expr) *Type {

	// TODO: build out expression type evaluation
	// using the expresssion list below

	switch t := e.(type) {
	case *ast.ParenExpr:
		// parenthetical expression (x)
		return f.TypeExpr(t.X)
	case *ast.Ident:
		// expression of true, false, iota, nil or a declared type
		return f.TypeIdent(t)
	case *ast.BasicLit:
		// literal expression of int, float, rune, string
		return TypeToken(t.Kind)
	case *ast.StarExpr:
		// pointer expression
		return f.TypePointer(f.TypeExpr(t.X))
	case *ast.SelectorExpr:
		// selector expression
	case *ast.UnaryExpr:
		// unary expression
		// TODO: t.X can be a type or value
		return f.TypeUnary(t.Op, f.TypeExpr(t.X))
	case *ast.BinaryExpr:
		// binary expression
	case *ast.CallExpr:
		switch t := t.Fun.(type) {
		case *ast.ParenExpr:
			return f.TypeExpr(t.X)

		}
	}
	fmt.Println("EXPR:", reflect.TypeOf(e))
	return nil

	/* switch t := v.Values[0].(type) {
	case *ast.TypeAssertExpr:
		// TODO: evaluate type assert expression
	case *ast.Ident:
		val.typ = TypeToken(token.IDENT)
	case *ast.BasicLit:
		val.typ = TypeToken(t.Kind)
	case *ast.UnaryExpr:
		// TODO: evaluate unary expression (&, *, etc.)
	case *ast.BinaryExpr:
		// TODO: evaluate binary expression
	case *ast.CallExpr:
		// TODO: evaluate call expression
	case *ast.FuncLit:
		// TODO: evaluate function literal
	case *ast.CompositeLit:
		switch t.Type.(type) {
		case *ast.Ident:
			val.typ = TypeToken(token.IDENT)
		case *ast.SelectorExpr:
			// TODO: evaluate selector expression
		case *ast.ArrayType:
			// TODO: evaluate array type
		case *ast.MapType:
			// TODO: evaluate map type
		case *ast.StructType:
			// TODO: evaluate struct type
		case *ast.InterfaceType:
			// TODO: evaluate interface type
		case *ast.ChanType:
			// TODO: evaluate chan type
		default:
			// TODO: check for other types
		}
	} */
}

// TypeIdent returns the type of the identifier in the file
// by checking builtin values and then declared types.
// TODO: Ident can be a type, value or function.
func (f *File) TypeIdent(i *ast.Ident) (typ *Type) {
	if v := BuiltinValues.Get(i.Name); v != nil {
		return v.(*Value).typ
	}
	if typ, _ = f.GetType(i.Name); typ != nil {
		return
	}
	fmt.Println("Unknown IDENT:", i.Name)
	if i.Obj != nil {
		switch i.Obj.Kind {
		case ast.Var:
			f.InspectValue(VAR, i.Obj.Decl.(*ast.ValueSpec))
			return f.p.Values.Get(i.Name).(*Value).typ
		case ast.Con:
			f.InspectValue(CONST, i.Obj.Decl.(*ast.ValueSpec))
			return f.p.Values.Get(i.Name).(*Value).typ
		case ast.Typ:
			f.InspectType([]ast.Spec{i.Obj.Decl.(*ast.TypeSpec)})
			return f.p.Types.Get(i.Name).(*Type)
		case ast.Fun:
			f.InspectFunc(i.Obj.Decl.(*ast.FuncDecl))
			// TODO: need to know the value placement to know
			// which output to use from the function
			//return f.p.Funcs.Get(i.Name).(*Func)
		}
		fmt.Println("OBJECT:",
			i.Obj.Kind,
			i.Obj.Name,
			reflect.TypeOf(i.Obj.Decl),
			reflect.TypeOf(i.Obj.Type),
		)
	}
	os.Exit(1)
	return
}

// TypePointer returns the pointer type of the type provided.
func (f *File) TypePointer(t *Type) (typ *Type) {
	n := "*" + t.name
	if t := f.p.Types.Get(n); t != nil {
		return t.(*Type)
	}
	typ = &Type{
		file:   t.file,
		name:   "*" + t.name,
		kind:   POINTER,
		object: &Pointer{elem: t},
	}
	typ.object.(*Pointer).typ = typ
	f.p.Types.Add(typ)
	return
}

// TypeUnary returns the unary type of the type provided.
func (f *File) TypeUnary(op token.Token, t *Type) (typ *Type) {
	switch op {
	case token.AND:
		return f.TypePointer(t)
	}
	return
}
