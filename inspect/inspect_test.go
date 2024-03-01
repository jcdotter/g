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
)

func TestAst(t *testing.T) {
	var err error
	p := NewPackage("github.com/jcdotter/go/data")
	p.Name = "data"
	f := NewFile(p, "data")
	f.t, err = parser.ParseFile(token.NewFileSet(), "../data/data.go", nil, 0)
	if err != nil {
		return
	}
	f.Inspect()
}

func TestNewPackage(t *testing.T) {
	NewPackage("github.com/jcdotter/go/data")
	NewPackage("strings")
	NewPackage("github.com/jackc/pgx/v5")
	/* _ = pgx.ErrNoRows
	_ = gotype.STRING("") */

}
