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
	"os"
	"path/filepath"
	"strings"

	"github.com/jcdotter/go/stack"
)

var SrcPath, PkgPath, ModPath, Mod string

var (
	Sep    = string(os.PathSeparator)
	ExtSep = "."
	UpDir  = ".."
)

func init() {
	// set the package path
	if p := os.Getenv("GOPATH"); p != "" {
		PkgPath = Join(p, "pkg", "mod")
	} else {
		PkgPath = Join(os.Getenv("HOME"), "go", "pkg", "mod")
	}
	// set the source path
	SrcPath = Join("/usr", "local", "go", "src")
	// set the module path
	if p, err := os.Getwd(); err == nil {
		Mod, _, ModPath, _ = GetModule(p)
	}
}

type Path struct {
	path  string
	abs   bool
	frame *stack.Frame
	mod   *Element // the path module
	pkg   *Element // the path package
	dir   *Element // the path directory
	file  *Element // the path file
	fn    *Element // the path function
}

// Exists returns true if the path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MustExist panics if the path does not exist
func MustExist(path string) {
	if !Exists(path) {
		panic(fmt.Errorf("path %s does not exist", path))
	}
}

// New returns a new Path
func New(path string) *Path {
	return &Path{
		path: path,
	}
}

// FilePath returns the path of a file
// at the given offset in stacktrace
func FilePath(skip int) *Path {
	f := stack.Caller(skip + 1)
	return &Path{
		frame: f,
		path:  f.Frame().File,
		abs:   true,
	}
}

// DirPath returns the path of a directory
// at the given offset in stacktrace
func DirPath(skip int) *Path {
	f := stack.Caller(skip + 1)
	return &Path{
		frame: f,
		path:  File(f.Frame().File).Dir,
		abs:   true,
	}
}

// CurrentFile returns the current file path
func CurrentFile() *Path {
	return FilePath(1)
}

// CurrentDir returns the current directory path
func CurrentDir() *Path {
	return DirPath(1)
}

func WorkingDir() *Path {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &Path{
		path: dir,
		abs:  true,
	}
}

// Set creates a path from the given 3 elements
// of dir, file, extension
func Set(dir, name, ext string) *Path {
	return New(dir + Sep + name + ExtSep + ext)
}

// IsAbs returns true if the path is absolute
func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Abs returns the absolute path
func Abs(rel string) string {
	p, err := filepath.Abs(rel)
	if err != nil {
		panic(err)
	}
	return p
}

// Clean returns the cleaned path
func Clean(path string) string {
	return filepath.Clean(path)
}

// Join joins the given paths
func Join(paths ...string) string {
	return filepath.Join(paths...)
}

// IsSet returns true if the path is set
func (p *Path) IsSet() bool {
	return p.path != ""
}

// IsAbs returns true if the path is absolute
func (p *Path) IsAbs() bool {
	return p.abs
}

// IsFile returns true if the path is a file
func (p *Path) IsFile() bool {
	return !p.IsDir()
}

// IsDir returns true if the path is a directory
func (p *Path) IsDir() bool {
	return IsDir(p.path)
}

// Exists returns true if the path exists
func (p *Path) Exists() bool {
	return Exists(p.path)
}

// SetPath sets the path
func (p *Path) SetPath(path string) *Path {
	p = p.Reset()
	p.path = path
	return p
}

// SetAbs updates the path to the absolute path
func (p *Path) SetAbs() *Path {
	if p.IsSet() && !p.abs {
		p.path = Abs(p.path)
		p.abs = true
	}
	return p
}

// SeRel returns the path relative to the given path
func (p *Path) SetRel(to string) *Path {
	path := p.Rel(to)
	mod := *p.mod
	p.Reset()
	p.path = path
	p.mod = &mod
	return p
}

// Parent returns the parent path
func (p *Path) Parent() *Path {
	if p.IsSet() {
		if i := strings.LastIndex(p.path, Sep); i != -1 {
			n := New(p.path[:i])
			if n.Dir().Name == UpDir {
				n.SetAbs()
			}
			return n
		}
	}
	return p
}

// Child returns the child path
func (p *Path) Child(child string) *Path {
	if p.IsSet() {
		if p.IsFile() {
			panic("cannot get child of file")
		}
		return New(p.path + Sep + child)
	}
	return p
}

// Build builds the path elements
func (p *Path) Build() *Path {
	if p.IsSet() {
		p.mod = Module(p.path)
		p.file = File(p.path)
		if p.frame != nil {
			p.pkg = Package(p.frame.Frame().Function)
			p.fn = Function(p.frame.Frame().Function)
		}
	}
	return p
}

// Reset resets the path
func (p *Path) Reset() *Path {
	p.path = ""
	p.abs = false
	p.frame = nil
	p.mod = nil
	p.pkg = nil
	p.file = nil
	p.fn = nil
	return p
}

// Path returns the path
func (p *Path) Path() string {
	return p.path
}

// Frame returns the frame
func (p *Path) Frame() *stack.Frame {
	return p.frame
}

// Module returns the module element
func (p *Path) Module() *Element {
	if p.IsSet() && p.mod == nil {
		p.mod = Module(p.path)
	}
	return p.mod
}

// Package returns the package element
func (p *Path) Package() *Element {
	if p.IsSet() && p.pkg == nil && p.frame != nil {
		p.pkg = Package(p.frame.Frame().Function)
	}
	return p.pkg
}

// Dir returns the directory element
func (p *Path) Dir() *Element {
	if p.IsSet() && p.dir == nil {
		p.dir = Directory(p.path)
	}
	return p.dir
}

// File returns the file element
func (p *Path) File() *Element {
	if p.IsSet() && p.file == nil {
		p.file = File(p.path)
	}
	return p.file
}

// Func returns the func element
func (p *Path) Func() *Element {
	if p.IsSet() && p.fn == nil && p.frame != nil {
		p.fn = Function(p.frame.Frame().Function)
	}
	return p.fn
}

// Abs returns the absolute path
func (p *Path) Abs() string {
	if p.IsSet() {
		if p.abs {
			return p.path
		}
		return Abs(p.path)
	}
	return p.path
}

// Rel returns the path relative to the given path
func (p *Path) Rel(to string) string {
	if p.IsSet() {
		to = Abs(to)
		path := p.Abs()
		fmt.Println(path, to)
		rel, err := filepath.Rel(to, path)
		if err != nil {
			panic(err)
		}
		return rel
	}
	return Abs(to)
}

// Files returns the files in the path
func (p *Path) Files() []string {
	if p.IsSet() {
		if p.IsFile() {
			return []string{p.path}
		}
		return Files(p.path)
	}
	return nil
}

// Files returns the files in the path
func Files(path string) (files []string) {
	if IsFile(path) {
		return []string{path}
	}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files
}
