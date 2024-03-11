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
	"os"
	"strings"

	"github.com/jcdotter/go/parser"
)

var modFile = "go.mod"

// Element represents a Path element such as
// a package, module, or file
type Element struct {
	Name string // the short name of the element
	Dir  string // the directory of the element
	Path string // the full path of the element
}

// Function returns a new function element from a function path
func Function(path string) *Element {
	_, i := parseFuncIndex(path)
	return &Element{
		Name: path[i+1:],
		Dir:  path[:i],
		Path: path,
	}
}

func parseFuncIndex(fn string) (i1, i2 int) {
	i1 = strings.LastIndex(fn, Sep) + 1
	i2 = i1 + strings.Index(fn[i1:], ExtSep)
	return
}

// File returns a new file element from a file path
func File(path string) *Element {
	if !IsFile(path) {
		return nil
	}
	i := strings.LastIndex(path, Sep)
	return &Element{
		Name: path[i+1:],
		Dir:  path[:i],
		Path: path,
	}
}

// IsFile returns true if the path is a file
func IsFile(path string) bool {
	return !IsDir(path)
}

// Directory returns a new directory element from a directory path
func Directory(path string) *Element {
	if !IsDir(path) {
		path = path[:strings.LastIndex(path, Sep)]
	}
	return &Element{
		Name: path[strings.LastIndex(path, Sep)+1:],
		Dir:  path,
		Path: path,
	}
}

// IsDir returns true if the path is a directory
func IsDir(path string) bool {
	i := strings.LastIndex(path, Sep)
	switch {
	case i == -1:
		return false
	case i == len(path)-1:
		return true
	default:
		return !strings.Contains(path[i+1:], ExtSep)
	}
}

// Package returns a new package element from a funcation path
func Package(path string) *Element {
	i1, i2 := parseFuncIndex(path)
	return &Element{
		Name: path[i1:i2],
		Dir:  path[:i1-1],
		Path: path[:i2],
	}
}

// Module returns a new module element frm a file or directory path
func Module(path string) *Element {
	m, _, d, p := GetModule(path)
	return &Element{
		Name: m,
		Dir:  d,
		Path: p,
	}
}

// GetFilePath searches for the given file in the current
// and each parent directory in the provided path and returns
// the directory and full path of the first file found
func GetFileUpPath(path, file string) (found bool, dir, fullpath string) {
	// get origin path
	if dir = GetAbsDir(path); dir == "" {
		return
	}
	// find file name and set mod path
	dir += Sep
	for i := len(dir) - 1; i > 0; i-- {
		if dir[i] == os.PathSeparator {
			dir = dir[:i]
			fullpath = dir + Sep + file
			if _, err := os.Stat(fullpath); !os.IsNotExist(err) {
				found = true
				return
			}
		}
	}
	return false, "", ""
}

// GetFilePath searches for the given file in the current
// and each child directory in the provided path and returns
// the directory and full path of the first file found
func GetFileDownPath(path, file string) (found bool, dir, fullpath string) {
	// get origin path
	if dir = GetAbsDir(path); dir == "" {
		return
	}
	return getFileDownPath(dir, file)
}

func getFileDownPath(path, file string) (found bool, dir, fullpath string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	dirs := make([]string, 0, 8)
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, path+Sep+entry.Name())
		} else if entry.Name() == file {
			found = true
			fullpath = path + Sep + file
			return
		}
	}
	for _, d := range dirs {
		if found, dir, fullpath = getFileDownPath(d, file); found {
			return
		}
	}
	return
}

// GetAbsDir returns the absolute path of the given path
// and returns an empty string if the path does not exist
func GetAbsDir(path string) (dir string) {
	switch info, err := os.Stat(path); {
	case os.IsNotExist(err):
		return
	case !info.IsDir():
		dir = Abs(path)
		return dir[:strings.LastIndex(dir, Sep)]
	default:
		return Abs(path)
	}
}

// GetModule searches for the go.mod file in the
// current and each parent in the provided path
// and returns the module name, version, directory,
// and full path of the go.mod file
func GetModule(modPath string) (mod, ver, dir, mpath string) {
	var found bool
	found, dir, mpath = GetFileUpPath(modPath, modFile)
	if !found {
		return
	}
	n := 0
	b, _ := os.ReadFile(mpath)
	m, n := parser.Module(b, n)
	v, _ := parser.Version(b, n)
	mod = string(m)
	ver = string(v)
	return
}

func GetPackagePath(fullName string) (pkg string) {
	// check working directory for the package
	if strings.HasPrefix(fullName, Mod) {
		return Join(ModPath, strings.TrimPrefix(fullName, Mod))
	}
	// check go src directory for the package
	pkg = Join(SrcPath, fullName)
	if f, err := os.Open(pkg); !os.IsNotExist(err) {
		defer f.Close()
		return
	}
	// check go mod directory for the package
	b, _ := os.ReadFile(Join(ModPath, modFile))
	if found, at := parser.Search([]byte(fullName), b, 0); found {
		_, vbeg := parser.Find('v', b, at+len(fullName))
		_, vend := parser.Next(parser.NOT(parser.IsChar), b, vbeg)
		pkg = Join(PkgPath, fullName) + "@" + string(b[vbeg:vend])
		if f, err := os.Open(pkg); !os.IsNotExist(err) {
			defer f.Close()
			return
		}
	}
	return ""
}
