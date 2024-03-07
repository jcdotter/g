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
	"go/parser"
	"go/token"
	"testing"

	"github.com/jcdotter/go/data"
	"github.com/jcdotter/go/path"
	"github.com/jcdotter/go/test"
)

var config = &test.Config{
	Trace:   true,
	Detail:  true,
	Require: true,
	Msg:     "%s",
}

// InspectFile is a test utiiity
func InspectFile(file string) *File {
	p := &Package{
		Name:    file,
		Path:    file,
		Imports: data.Make[*Package](data.Cap),
		Files:   data.Make[*File](data.Cap),
		Values:  data.Make[*Value](data.Cap),
		Types:   data.Make[*Type](data.Cap),
		Funcs:   data.Make[*Func](data.Cap),
	}
	f := NewFile(p, file)
	p.Files.Add(f)
	f.t, _ = parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
	f.Capture()
	f.Inspect()
	return f

}

func TestNewPackage(t *testing.T) {
	gt := test.New(t, config)
	var n string
	var p *Package

	n = "github.com/jcdotter/go/test"
	gt.Msg = "NewPackage(\"%s\").%s"
	p = NewPackage(n)
	gt.Equal(n, p.Name, n, "Name")
	gt.Equal(path.Abs("../test"), p.Path, n, "Path")

	n = "strings"
	gt.Msg = "NewPackage(\"%s\").%s"
	p = NewPackage(n)
	gt.Equal(n, p.Name, n, "Name")
	gt.Equal(SrcPath+n, p.Path, n, "Path")

	n = "golang.org/x/term@v0.17.0"
	gt.Msg = "NewPackage(\"%s\").%s"
	p = NewPackage(n)
	gt.Equal(n, p.Name, n, "Name")
	gt.Equal(PkgPath+n, p.Path, n, "Path")
}

func TestInspect(t *testing.T) {
	gt := test.New(t, config)
	f := InspectFile("sample_test.go")
	/* var e data.Elem

	// InspectImports
	gt.Msg = "Inspect().Imports.%s"
	gt.Equal(1, f.i.Len(), "Len")
	e = f.i.Get("_")
	gt.NotNil(e, "Get")
	i := e.(*Import)
	gt.Equal(f, i.File(), "File")
	gt.Equal("_", i.Name(), "Name")
	gt.Equal("testing", i.pkg.Name, "pkg.Name")
	gt.Equal(SrcPath+"testing", i.pkg.Path, "pkg.Path") */

	// InspectValues
	var c *Value

	// InspectConsts
	gt.Msg = "Inspect().Const.%s"
	c = f.p.Values.Get("Const0").(*Value)
	gt.Equal(f, c.file, "File")
	gt.Equal("Const0", c.Name(), "Name")
	gt.Equal(CONST, c.Kind(), "const/var")
	gt.Equal(INT, c.Type().Kind(), "Kind")
	gt.Equal(BuiltinTypes.Get("int"), c.typ, "Type")

	c = f.p.Values.Get("Const1").(*Value)
	gt.Equal(INT, c.typ.Kind(), "Kind")

	c = f.p.Values.Get("Const2").(*Value)
	gt.Equal(BYTE, c.typ.Kind(), "Kind")

	c = f.p.Values.Get("Const3").(*Value)
	gt.Equal(BYTE, c.Type().Kind(), "Kind")

	// InspectBasicLits
	gt.Msg = "Inspect().BasicLits.%s"
	c = f.p.Values.Get("IntBasic").(*Value)
	gt.Equal(VAR, c.Kind(), "const/var")
	gt.Equal(INT, c.Type().Kind(), "Kind")

	c = f.p.Values.Get("FloatBasic").(*Value)
	gt.Equal(FLOAT64, c.Type().Kind(), "Kind")

	c = f.p.Values.Get("ComplexBasic").(*Value)
	gt.Equal(COMPLEX128, c.Type().Kind(), "Kind")

	c = f.p.Values.Get("StringBasic").(*Value)
	gt.Equal(STRING, c.Type().Kind(), "Kind")

	c = f.p.Values.Get("RuneBasic").(*Value)
	gt.Equal(RUNE, c.Type().Kind(), "Kind")

	// InspectParenExprs
	gt.Msg = "Inspect().Paren.%s"
	c = f.p.Values.Get("IntParen").(*Value)
	gt.Equal(INT, c.Type().Kind(), "Kind")

	// InspectPointerTypes
	gt.Msg = "Inspect().Pointer.%s"
	c = f.p.Values.Get("IntPointer").(*Value)
	gt.Equal(POINTER, c.Type().Kind(), "Kind")
	gt.Equal(INT, c.Type().Object().(*Pointer).Elem().Kind(), "Elem.Kind")

	// InspectPointerRefs
	gt.Msg = "Inspect().PointerRef.%s"
	c = f.p.Values.Get("IntRef").(*Value)
	gt.Equal(POINTER, c.Type().Kind(), "Kind")
	gt.Equal(INT, c.Type().Object().(*Pointer).Elem().Kind(), "Elem.Kind")

	// inspectBinaryExprs
	gt.Msg = "Inspect().Binary.%s"
	c = f.p.Values.Get("StringBinary").(*Value)
	gt.Equal(STRING, c.Type().Kind(), "Kind")

	// InspectCallExprs
	gt.Msg = "Inspect().Call.%s"
	c = f.p.Values.Get("IntCall").(*Value)
	gt.Equal(INT, c.Type().Kind(), "Kind")

	// InspectFuncLits
	gt.Msg = "Inspect().FuncLit.%s"
	c = f.p.Values.Get("FuncLit").(*Value)
	gt.Equal(FUNC, c.Type().Kind(), "Kind")
	gt.Equal(0, c.Type().Object().(*Func).In().Len(), "NumParams")
	gt.Equal(1, c.Type().Object().(*Func).Out().Len(), "NumResults")
	gt.Equal(INT, c.Type().Object().(*Func).Out().Index(0).(*Type).Kind(), "Result.Kind")

	// InspectArrayLits
	gt.Msg = "Inspect().ArrayLit.%s"
	c = f.p.Values.Get("ArrayLit").(*Value)
	gt.Equal(ARRAY, c.Type().Kind(), "Kind")
	gt.Equal(INT, c.Type().Object().(*Array).Elem().Kind(), "Elem.Kind")

	// InspectSliceLits
	gt.Msg = "Inspect().SliceLit.%s"
	c = f.p.Values.Get("SliceLit").(*Value)
	gt.Equal(SLICE, c.Type().Kind(), "Kind")
	gt.Equal(INT, c.Type().Object().(*Array).Elem().Kind(), "Elem.Kind")

	// InspectMapLits
	gt.Msg = "Inspect().MapLit.%s"
	c = f.p.Values.Get("MapLit").(*Value)
	gt.Equal(MAP, c.Type().Kind(), "Kind")
	gt.Equal(STRING, c.Type().Object().(*Map).Key().Kind(), "Key.Kind")
	gt.Equal(INT, c.Type().Object().(*Map).Elem().Kind(), "Elem.Kind")

	// InspectChanLits
	gt.Msg = "Inspect().ChanLit.%s"
	c = f.p.Values.Get("ChanLit").(*Value)
	gt.Equal(CHAN, c.Type().Kind(), "Kind")
	gt.Equal(STRING, c.Type().Object().(*Chan).Elem().Kind(), "Elem.Kind")

	// InspectStructLits
	gt.Msg = "Inspect().StructLit.%s"
	c = f.p.Values.Get("StructLit").(*Value)
	gt.Equal(STRUCT, c.Type().Kind(), "Kind")
	gt.Equal(STRING, c.Type().Object().(*Struct).Index(1).Type().Kind(), "Field.Kind")

	// InspectSelExprs
	gt.Msg = "Inspect().SelExpr.%s"
	c = f.p.Values.Get("SelExpr").(*Value)
	gt.Equal(STRING, c.Type().Kind(), "Kind")

	// InspectIndexExprs
	gt.Msg = "Inspect().IndexExpr.%s"
	c = f.p.Values.Get("IndexExpr").(*Value)
	gt.Equal(INT, c.Type().Kind(), "Kind")

	// InspectTypes
	var x *Type

	// InspectBoolType
	gt.Msg = "Inspect().BoolType.%s"
	x = f.p.Types.Get("BoolType").(*Type)
	gt.Equal(BOOL, x.Kind(), "Kind")
	gt.Equal(POINTER, x.Methods().Index(0).(*Func).Of().Kind(), "Method.Receiver.Kind")
	gt.Equal("BoolType", x.Methods().Index(0).(*Func).Of().Object().(*Pointer).Elem().Name(), "Method.Receiver.Element.Name")
	gt.Equal(INT, x.Methods().Index(0).(*Func).In().Index(0).(*Type).Kind(), "Method.Param.Kind")
	gt.Equal(INT, x.Methods().Index(0).(*Func).Out().Index(0).(*Type).Kind(), "Method.Result.Kind")

}
