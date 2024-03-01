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
	"go/parser"
	"go/token"
	"testing"

	"github.com/jcdotter/go/data"
)

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

func TestInspect(t *testing.T) {
	Inspect("github.com/jcdotter/go/data")
}

func TestAst(t *testing.T) {
	f := InspectFile("sample_test.go")
	fmt.Println("COMPLETED:", f.n)
}

func TestNewPackage(t *testing.T) {
	NewPackage("github.com/jcdotter/go/data")
	NewPackage("strings")
	NewPackage("github.com/jackc/pgx/v5")
	/* _ = pgx.ErrNoRows
	_ = gotype.STRING("") */

}
