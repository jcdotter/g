// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/grpg/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inspect

import (
	"go/ast"
	"go/parser"
	"go/token"
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
	return p, nil
}

// Parse parses the package content if not already parsed. If Entites are provided,
// the package will only parse the provided entities, otherwise the package will
// parse all entities in the package. Returns an error if the package cannot be parsed.
// TODO: Make file parsing concurrent.
func (p *Package) Parse() (err error) {
	fset := token.NewFileSet()

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
		file.t, err = parser.ParseFile(fset, f, nil, parser.ParseComments)
		if err != nil {
			return
		}
	}
	return
}

func (f *File) Inspect() (err error) {
	ast.Inspect(f.t, func(n ast.Node) bool {

		return true
	})
	return
}
