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

package path

import (
	"fmt"
	"testing"
)

var (
	path = Abs("../path/path.go")
	mod  = Abs("../../go.mod")
)

func TestMod(t *testing.T) {
	fmt.Println(IsFile("/home/decodex/lib/grpg/generator/helps.go"))
	fmt.Println(IsDir("/home/decodex/lib/grpg/generator/helps.go"))
	fmt.Println(New("../../go.mod").Path())
	fmt.Println(New("../../go.mod").Parent().Path())
}

func BenchmarkAbs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Abs("../path/path.go")
	}
}

func BenchmarkPathUp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFileUpPath(path, "go.mod")
	}
}

func BenchmarkPathDown(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFileDownPath(mod, "path.go")
	}
}

func BenchmarkModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Module(path)
	}
}

func BenchmarkPackage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Package(path)
	}
}

func BenchmarkDirectory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Directory(path)
	}
}

func BenchmarkFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		File(path)
	}
}

func BenchmarkFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Function(path)
	}
}
