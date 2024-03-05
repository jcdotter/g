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
	"strconv"

	"github.com/jcdotter/go/data"
	"github.com/jcdotter/go/path"
)

// Inspect parses the package content in the path provided and
// returns the package object for inspection, or an error if
// the package cannot be parsed.
func Inspect(PkgPath string) (*Package, error) {
	p := NewPackage(PkgPath)
	if err := p.Inspect(); err != nil {
		return nil, err
	}
	return p, nil
}

// Inspect parses the package content and
// returns the package object for inspection, or an error if
// the package cannot be parsed.
func (p *Package) Inspect() (err error) {
	fmt.Println("INSPECTING PACKAGE:", p.Name)
	if err = p.Parse(); err != nil {
		return
	}
	for _, f := range p.Files.List() {
		fmt.Println("  CAPTURING FILE:", f.(*File).n)
		if err = f.(*File).Capture(); err != nil {
			return
		}
	}
	for _, f := range p.Files.List() {
		fmt.Println("  INSPECTING FILE:", f.(*File).n)
		if err = f.(*File).Inspect(); err != nil {
			return
		}
	}
	p.Inspected = true
	return
}

// Parse parses the package content if not already parsed. If Entites are provided,
// the package will only parse the provided entities, otherwise the package will
// parse all entities in the package. Returns an error if the package cannot be parsed.
func (p *Package) Parse() (err error) {
	// TODO: Make file parsing concurrent.

	// parse each file in the package
	for _, f := range path.Files(p.Path) {
		if IsGoFile(f) {
			var file *File

			// parse file name
			n := f[:len(f)-3]

			// check if file is already parsed
			// else add a new file to the package
			if f := p.Files.Get(n); f != nil {
				return
			}
			file = NewFile(p, n)
			p.Files.Add(file)
			f = p.Path + string(os.PathSeparator) + f
			// parse file to abstract syntax tree
			file.t, err = parser.ParseFile(token.NewFileSet(), f, nil, parser.SkipObjectResolution)
			if err != nil {
				return
			}
		}
	}

	// parse package name
	if p.Files.Len() > 0 {
		p.Name = p.Files.Index(0).(*File).t.Name.Name
	}

	return
}

// TypeIdent returns the type of the identifier in the package
func (p *Package) TypeIdent(ident string) (typ *Type) {
	// parse if package has not been parsed
	if !p.Inspected && p.Inspect() != nil {
		return
	}
	// check declared values, types, funcs for ident
	if v := p.Values.Get(ident); v != nil {
		return v.(*Value).typ
	}
	if t := p.Types.Get(ident); t != nil {
		return t.(*Type)
	}
	if f := p.Funcs.Get(ident); f != nil {
		return f.(*Func).typ
	}
	return
}

// Capture captures the declared entities in the file and
// adds them to the package.
func (f *File) Capture() (err error) {
	if f.t != nil {
		// route declaration and capture
		// them into the package entities
		for _, d := range f.t.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				err = f.CaptureFunc(decl)
			case *ast.GenDecl:
				switch decl.Tok {
				case token.CONST:
					err = f.CaptureValues(CONST, decl.Specs)
				case token.VAR:
					err = f.CaptureValues(VAR, decl.Specs)
				case token.TYPE:
					err = f.CaptureTypes(decl.Specs)
				case token.IMPORT:
					err = f.CaptureImports(decl.Specs)
				}
			}
		}
	}
	return
}

// CaptureImports captures the import declarations
// in the file and adds them to the file.
func (f *File) CaptureImports(specs []ast.Spec) (err error) {
	for _, s := range specs {
		i := s.(*ast.ImportSpec)
		imp := &Import{file: f, spec: i}
		var path string
		imp.name, path = PackageName(i.Name, i.Path.Value)
		if imp.name != "_" && imp.name != "" {
			f.i.Add(imp)
		}
		// get package by path if already imported another file,
		// otherwise create a new imported package and add it to
		// the current package
		if pkg := f.p.Imports.Get(path); pkg != nil {
			imp.pkg = pkg.(*Package)
		} else {
			imp.pkg = NewPackage(path)
			imp.pkg.IsImport = true
			f.p.Imports.Add(imp.pkg)
		}
	}
	return
}

// CaptureValues captures the value declarations
// in the file and adds them to the package.
func (f *File) CaptureValues(k byte, specs []ast.Spec) (err error) {
	for _, s := range specs {
		var vals = s.(*ast.ValueSpec)
		for i, n := range vals.Names {
			f.p.Values.Add(
				&Value{
					file: f,
					kind: k,
					name: n.Name,
					spec: vals,
					indx: i,
				},
			)
		}
	}
	return
}

// CaptureTypes captures the type declarations
// in the file and adds them to the package.
func (f *File) CaptureTypes(specs []ast.Spec) (err error) {
	for _, s := range specs {
		t := s.(*ast.TypeSpec)
		f.p.Types.Add(
			&Type{
				file: f,
				name: t.Name.Name,
				spec: t,
			},
		)
	}
	return
}

// CaptureFunc captures a function declaration
// in the file and adds it to the package.
func (f *File) CaptureFunc(fn *ast.FuncDecl) (err error) {
	f.p.Funcs.Add(
		&Func{
			file: f,
			name: fn.Name.Name,
			spec: fn,
		},
	)
	return
}

func (f *File) Inspect() (err error) {
	if err = f.InspectValues(); err != nil {
		return err
	}
	if err = f.InspectTypes(); err != nil {
		return err
	}
	if err = f.InspectFuncs(); err != nil {
		return err
	}
	return nil
}

func (f *File) InspectValues() (err error) {
	var priorSpec *ast.ValueSpec
	var priorType *Type
	for i, vals := 0, f.p.Values.List(); i < len(vals); i++ {

		// assert value and skip if not in file
		v := vals[i].(*Value)
		if v.file != f {
			continue
		}

		// get declared type if it exists. if not,
		// the prior declared type will set the
		// type of this value if there are no
		// assginment expressions in this value.
		switch {
		case v.spec.Type != nil:
			v.typ = f.TypeExpr(v.spec.Type, nil)
			priorSpec, priorType = v.spec, v.typ
			continue
		case v.spec == priorSpec && priorType != nil:
			v.typ = priorType
			continue
		}

		priorType = nil
		v.typ = f.TypeExpr(v.spec.Values[v.indx], nil)

		// if the type is not declared and the value
		// is function call, iterate through the output
		// types and add them to the value types
		if len(v.spec.Values) == 1 && v.typ != nil && v.typ.kind == FUNC && v.typ.object != nil {
			l := v.typ.object.(*Func).out.Len() - 1
			for j, t := range v.typ.object.(*Func).out.List() {
				v := vals[i+j].(*Value)
				if v.indx == j {
					v.typ = t.(*Type)
				}
			}
			i += l
		}
	}
	return
}

func (f *File) InspectTypes() (err error) {
	for _, x := range f.p.Types.List() {

		// assert type and skip if not in file
		t := x.(*Type)
		if t.file != f {
			continue
		}

		// inspect type expression and update
		// type with kind and object
		typ := f.TypeExpr(t.spec.Type, t.spec.TypeParams)
		t.kind = typ.kind
		t.object = typ.object
		fmt.Println("INSPECT TYPE:", t.name)
	}
	return
}

func (f *File) InspectFuncs() (err error) {
	for _, x := range f.p.Funcs.List() {

		// assert func and skip if not in file
		fn := x.(*Func)
		if fn.file != f {
			continue
		}

		// create new type using function spec type
		// and replace the package function with the
		// new function type
		typ := f.TypeFunc(fn.spec.Type, nil)
		fnc := typ.object.(*Func)
		fnc.name = fn.name
		fnc.spec = fn.spec
		f.p.Funcs.Set(fn.name, fnc)

		// if received has one identifier, inspect
		// the receiver, add it to the function, and
		// add the function as a method to the receiver
		if fn.spec.Recv != nil {
			if len(fn.spec.Recv.List) == 1 {
				//fmt.Println("RECEIVER:", fn.spec.Recv.List[0].Names[0].Name, fn.spec.Recv.List[0].Type, fn.name)
				if i := fn.spec.Recv.List[0].Names; len(i) == 1 {
					if rtyp := f.TypeExpr(fn.spec.Recv.List[0].Type, nil); rtyp != nil {
						fn.of = rtyp
						if rtyp.kind == POINTER {
							rtyp = rtyp.object.(*Pointer).elem
						}
						rtyp.methods.Add(fn)
					}
				}
			}
		}
	}
	return
}

// ----------------------------------------------------------------------------
// Type Evaluation Methods

func (f *File) TypeExpr(e ast.Expr, p *ast.FieldList) *Type {
	// TODO: evaluate need for the following expressions:
	// *ast.IndexExpr:
	// *ast.SliceExpr:
	fmt.Println("EXPR:", reflect.TypeOf(e))
	switch t := e.(type) {
	case *ast.BasicLit:
		// literal expression of int, float, rune, string
		return TypeToken(t.Kind)
	case *ast.ParenExpr:
		return f.TypeExpr(t.X, p)
	case *ast.Ident:
		return f.TypeIdent(t, p)
	case *ast.StarExpr:
		return f.TypePointer(t, p)
	case *ast.UnaryExpr:
		return f.TypeUnary(t, p)
	case *ast.BinaryExpr:
		return f.TypeBinary(t, p)
	case *ast.CallExpr:
		return f.TypeExpr(t.Fun, p)
	case *ast.FuncLit:
		return f.TypeFuncLit(t, p)
	case *ast.FuncType:
		return f.TypeFunc(t, p)
	case *ast.CompositeLit:
		return f.TypeExpr(t.Type, p)
	case *ast.ArrayType:
		return f.TypeArray(t, p)
	case *ast.MapType:
		return f.TypeMap(t, p)
	case *ast.StructType:
		return f.TypeStruct(t, p)
	case *ast.InterfaceType:
		return f.TypeIterface(t, p)
	case *ast.ChanType:
		return f.TypeChan(t, p)
	case *ast.SelectorExpr:
		return f.TypeSelector(t, p)
	case *ast.TypeAssertExpr:
		return f.TypeExpr(t.Type, p)
	}
	fmt.Println("UNKNOWN EXPR:", reflect.TypeOf(e))
	return nil
}

// TypeIdent returns the type of the identifier in the file
// by first checking builtin values, then declared types,
// and lastly inspecting the ident object if it exists.
func (f *File) TypeIdent(i *ast.Ident, x *ast.FieldList) (typ *Type) {

	// check decl spec type if it exists
	if x != nil {
		fmt.Println("PARAMS:", len(x.List))
		for _, field := range x.List {
			if len(field.Names) == 0 {
				return f.TypeExpr(field.Type, x)
			}
			for _, n := range field.Names {
				if n.Name == i.Name {
					fmt.Println("PARAM:", f.TypeExpr(field.Type, x))
					return f.TypeExpr(field.Type, x)
				}
			}
		}
	}

	// check builtin values, types, funcs
	if v := BuiltinValues.Get(i.Name); v != nil {
		return v.(*Value).typ
	}
	if t := BuiltinTypes.Get(i.Name); t != nil {
		return t.(*Type)
	}
	if f := BuiltinFuncs.Get(i.Name); f != nil {
		return f.(*Func).typ
	}

	// check declared values, types, funcs
	if v := f.p.Values.Get(i.Name); v != nil {
		return v.(*Value).typ
	}
	if t := f.p.Types.Get(i.Name); t != nil {
		return t.(*Type)
	}
	if f := f.p.Funcs.Get(i.Name); f != nil {
		return f.(*Func).typ
	}

	// if ident is an import, parse the imported package
	// and return teh imported package as a type
	if imp := f.i.Get(i.Name); imp != nil {
		typ = &Type{
			file: f,
			kind: PACKAGE,
			imp:  imp.(*Import),
			name: imp.(*Import).name,
		}
		// only inspect the first depth level of imported packages,
		// for deeper dependencies, inspect the package when a
		// declaration from the package is encountered. See
		// TypeSelector and TypeIndex for more reference.
		//if !f.p.IsImport {
		if p := imp.(*Import).pkg; p.Inspected || p.Inspect() == nil {
			typ.object = p
		}
		//}
		return
	}
	fmt.Println("IDENT NOT FOUND:", f.p.Name+"."+i.Name, "  OBJ:", i.Obj)
	fmt.Println("  PKG DECLS:", len(f.t.Decls), "CAPTURED:", f.p.Imports.Len()+f.p.Values.Len()+f.p.Types.Len()+f.p.Funcs.Len())
	for _, x := range f.p.Values.List() {
		v := x.(*Value)
		fmt.Println("  VALUE:", v.kind, v.name, v.typ)
	}
	for _, x := range f.p.Types.List() {
		t := x.(*Type)
		fmt.Println("  TYPE:", t.name, t.kind, t.object)
	}
	for _, x := range f.p.Funcs.List() {
		fn := x.(*Func)
		fmt.Println("  FUNC:", fn.name, fn.typ)
	}
	// if ident is not in types and has an object,
	// inspect the object and return the type
	/* if i.Obj != nil {
		switch i.Obj.Kind {
		case ast.Var:
			if err := f.InspectValue(VAR, i.Obj.Decl.(*ast.ValueSpec)); err == nil {
				return f.p.Values.Get(i.Name).(*Value).typ
			}
		case ast.Con:
			if err := f.InspectValue(CONST, i.Obj.Decl.(*ast.ValueSpec)); err == nil {
				return f.p.Values.Get(i.Name).(*Value).typ
			}
		case ast.Typ:
			if err := f.InspectType(i.Obj.Decl.(*ast.TypeSpec)); err == nil {
				return f.p.Types.Get(i.Name).(*Type)
			}
		case ast.Fun:
			if err := f.InspectFunc(i.Obj.Decl.(*ast.FuncDecl)); err == nil {
				return f.p.Funcs.Get(i.Name).(*Func).typ
			}
		}
	} */
	return
}

// TODO: rework if receiver changed to *Type from *File
// typePointer returns the pointer type of the type provided.
func (f *File) typePointer(t *Type) (typ *Type) {
	if t != nil {
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
	}
	return
}

// TypePointer returns the pointer type of the type provided.
func (f *File) TypePointer(p *ast.StarExpr, x *ast.FieldList) (typ *Type) {
	return f.typePointer(f.TypeExpr(p.X, x))
}

// TypeUnary returns the unary type of the type provided.
// Typically used for &Expr expressions.
func (f *File) TypeUnary(u *ast.UnaryExpr, x *ast.FieldList) (typ *Type) {
	switch u.Op {
	case token.AND:
		return f.typePointer(f.TypeExpr(u.X, x))
	}
	return
}

// TypeBinary returns the binary type of the types provided.
// Typically used for equation expressions (0+0 or ""+"").
func (f *File) TypeBinary(b *ast.BinaryExpr, x *ast.FieldList) (typ *Type) {
	// return type of the left expression if it is not
	// a basic literal. Go only adapts the type of basic
	// literals, so we can assume the type of a non-basic
	// literal is the type of the binary expression.
	e := b.X
	if p, ok := e.(*ast.ParenExpr); ok {
		e = p.X
	}
	if _, ok := e.(*ast.BasicLit); !ok {
		return f.TypeExpr(e, x)
	}
	return f.TypeExpr(b.Y, x)
}

// TypeFunc returns the type of the function literal provided.
// Typically used for assigning a function literal to a variable.
func (f *File) TypeFuncLit(fn *ast.FuncLit, x *ast.FieldList) (typ *Type) {
	// TODO: implement function literal as a function in the package.
	// currently only stored as a value in the package.
	fnc := &Func{file: f}
	typ = &Type{
		file:   f,
		kind:   FUNC,
		object: fnc,
	}
	fnc.typ = typ
	params := JoinFields(x, fn.Type.TypeParams)
	f.TypeParams(fn.Type.Params, fnc.in, params)
	f.TypeParams(fn.Type.Results, fnc.out, params)
	return
}

func JoinFields(a, b *ast.FieldList) *ast.FieldList {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return &ast.FieldList{List: append(a.List, b.List...)}
}

// TypeFunc returns the type of the function expression provided.
func (f *File) TypeFunc(fn *ast.FuncType, x *ast.FieldList) (typ *Type) {

	// set up function type
	fnc := &Func{
		file: f,
		in:   data.Make[*Type](4),
		out:  data.Make[*Type](4),
	}
	typ = &Type{
		file:   f,
		kind:   FUNC,
		object: fnc,
	}
	fnc.typ = typ
	params := JoinFields(x, fn.TypeParams)

	if fn.TypeParams != nil {
		for _, t := range fn.TypeParams.List {
			fmt.Println("TYPEPARAM:", t.Names[0].Name, t.Type)
		}
	}

	// build and add func inputs and outputs
	f.TypeParams(fn.Params, fnc.in, params)
	f.TypeParams(fn.Results, fnc.out, params)

	return
}

// TypeFuncParams adds the type of the function
// parameters provided to the data list provided.
func (f *File) TypeParams(from *ast.FieldList, to *data.Data, with *ast.FieldList) {
	if from != nil {
		for _, field := range from.List {
			t := f.TypeExpr(field.Type, with)
			if len(field.Names) == 0 {
				to.Add(t)
				continue
			}
			for range field.Names {
				to.Add(t)
			}
		}
	}
}

// TypeArray returns the type of the array expression provided.
func (f *File) TypeArray(a *ast.ArrayType, x *ast.FieldList) (typ *Type) {
	arr := &Array{elem: f.TypeExpr(a.Elt, x)}
	typ = &Type{
		file:   f,
		object: arr,
	}
	arr.typ = typ
	if a.Len == nil {
		typ.kind = SLICE
		typ.name = "[]" + arr.elem.name
	} else if _, ok := a.Len.(*ast.Ellipsis); ok {
		typ.kind = ELIPS
		typ.name = "..." + arr.elem.name
	} else if _, ok := a.Len.(*ast.BinaryExpr); ok {
		typ.kind = ARRAY
		typ.name = "[x]" + arr.elem.name
	} else {
		typ.kind = ARRAY
		l := a.Len.(*ast.BasicLit).Value
		arr.len, _ = strconv.Atoi(l)
		typ.name = "[" + l + "]" + arr.elem.name
	}
	return
}

// TypeMap returns the type of the map expression provided.
func (f *File) TypeMap(m *ast.MapType, x *ast.FieldList) (typ *Type) {
	mp := &Map{key: f.TypeExpr(m.Key, x), elem: f.TypeExpr(m.Value, x)}
	typ = &Type{
		file:   f,
		kind:   MAP,
		object: mp,
	}
	mp.typ = typ
	typ.name = "map[" + mp.key.name + "]" + mp.elem.name
	return
}

// TypeStruct returns the type of the struct expression provided.
func (f *File) TypeStruct(s *ast.StructType, x *ast.FieldList) (typ *Type) {

	// set up struct type
	typ = &Type{
		file: f,
		kind: STRUCT,
	}
	str := NewStruct(typ)
	typ.object = str

	// loop fields and add them to the struct
	if s.Fields != nil {
		offset := 0
		for _, field := range s.Fields.List {

			// type and tag are at a field scoping level
			// and may apply to multiple field names
			ftyp := f.TypeExpr(field.Type, x)
			ftag := ""
			if field.Tag != nil {
				ftag = field.Tag.Value
			}

			// add field to struct
			for _, n := range field.Names {
				str.fields.Add(&Field{
					name:   n.Name,
					of:     typ,
					typ:    ftyp,
					tag:    ftag,
					offset: offset,
				})
				offset++
			}
		}
	}
	return
}

// TypeIterface returns the type of the interface expression provided.
func (f *File) TypeIterface(i *ast.InterfaceType, x *ast.FieldList) (typ *Type) {

	// if the interface has no methods,
	// return an empty interface
	if i.Methods == nil {
		return BuiltinTypes.List()[INTERFACE].(*Type)
	}

	// set up interface type and add methods
	intr := &Interface{methods: data.Make[*Func](i.Methods.NumFields())}
	typ = &Type{
		file:   f,
		kind:   INTERFACE,
		object: intr,
	}
	intr.typ = typ
	for _, field := range i.Methods.List {

		// skip if the interface method has no name
		if len(field.Names) == 0 {
			continue
		}

		// if interface field is a func, add it to the interface
		if t := f.TypeExpr(field.Type, x); t != nil && t.kind == FUNC {
			if ftyp := f.TypeFunc(field.Type.(*ast.FuncType), x); ftyp != nil {
				ftyp.object.(*Func).name = field.Names[0].Name
				intr.methods.Add(ftyp.object.(*Func))
			}
		}
	}

	return
}

// TypeChan returns the type of the chan expression provided.
func (f *File) TypeChan(c *ast.ChanType, x *ast.FieldList) (typ *Type) {
	cn := &Chan{
		elem: f.TypeExpr(c.Value, x),
		dir:  byte(c.Dir),
	}
	typ = &Type{
		file:   f,
		kind:   CHAN,
		object: cn,
	}
	cn.typ = typ
	switch cn.dir {
	case SEND:
		typ.name = "chan<- " + cn.elem.name
	case RECV:
		typ.name = "<-chan " + cn.elem.name
	default:
		typ.name = "chan " + cn.elem.name
	}
	return
}

// TypeSelector returns the type of the selector expression provided.
// Typically used for call to an external package function, value or
// type or call to internal package method or struct field
func (f *File) TypeSelector(s *ast.SelectorExpr, x *ast.FieldList) (typ *Type) {
	// TODO: know if a func call or a func assigned to a variable
	// TODO: could have in indexExpr rather than selectorExpr

	// Type{file: current file, kind: package, object: imported package}
	//   |-- Type{file: import file, kind: struct}
	//   |-- Type{file: import file, kind: struct}
	//   |     |-- Field{of: Type, typ: Type}
	//   |     |-- Field{of: Type, typ: Type}

	idents := append(f.IdentExpr(s.X, x), s.Sel)
	if typ = f.TypeIdent(idents[0], x); typ != nil {
		for i := 1; i < len(idents); i++ {
			typ = f.TypeIdentKey(typ, idents[i].Name, x)
		}
	}
	f.PrintSelector(idents) // TODO: remove
	return
}

// TODO: remove
func (f *File) PrintSelector(i []*ast.Ident) {
	var n string
	for j, x := range i {
		if j > 0 {
			n += ":"
		}
		n += x.Name
		if x.Obj != nil && x.Obj.Decl != nil {
			n += "(" + reflect.TypeOf(x.Obj.Decl).String() + ")"
		}
	}
	fmt.Println("SELC EXPR:", n)
}

func (f *File) TypeIdentKey(t *Type, key string, x *ast.FieldList) (typ *Type) {
	if t != nil {
		if t.object != nil {
			switch t := t.object.(type) {
			case *Struct:
				return t.fields.Get(key).(*Field).typ
			case *Package:
				return t.TypeIdent(key)
			case *Func:
				i := &ast.Ident{Name: t.out.Index(0).Key()}
				tt := f.TypeIdent(i, x)
				return f.TypeIdentKey(tt, key, x)
			case *Map:
				return t.elem
			case *Array:
				return t.elem
			}
		}
		fmt.Println("TYPE:", t.name, "MISSING OBJECT")
		if t.kind == PACKAGE {
			return &Type{
				imp:  t.imp,
				name: key,
			}
		}
		if t.kind == 0 && t.imp != nil {
			return t.imp.pkg.TypeIdent(key)
		}
	}
	return
}

func (f *File) IdentExpr(e ast.Expr, x *ast.FieldList) (i []*ast.Ident) {
	switch t := e.(type) {
	case *ast.Ident:
		return []*ast.Ident{t}
	case *ast.SelectorExpr:
		return append(f.IdentExpr(t.X, x), t.Sel)
	case *ast.ParenExpr:
		return f.IdentExpr(t.X, x)
	case *ast.StarExpr:
		return f.IdentExpr(t.X, x)
	case *ast.UnaryExpr:
		return f.IdentExpr(t.X, x)
	case *ast.CallExpr:
		return f.IdentExpr(t.Fun, x)
	case *ast.TypeAssertExpr:
		return f.IdentExpr(t.Type, x)
	case *ast.CompositeLit:
		return f.IdentExpr(t.Type, x)
	case *ast.ArrayType:
		return f.IdentExpr(t.Elt, x)
	case *ast.MapType:
		return f.IdentExpr(t.Value, x)
	case *ast.ChanType:
		return f.IdentExpr(t.Value, x)
	}
	return
}
