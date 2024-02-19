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

package stack

import (
	"runtime"
	"strings"
)

// Frame is a record from the stacktrace
type Frame struct {
	stack *Stack
	frame *runtime.Frame
	pkg   *src // stacktrace frame package
	file  *src // stacktrace frame file
	fn    *src // stacktrace frame function
	line  int  // stacktrace frame file line
}

// src is an element from a stacktrace frame
type src struct {
	Name string // the short name of the element
	Dir  string // the directory of the element
	Path string // the full path of the element
}

// Caller returns a frame from the stacktrace,
// skipping the number of frames specified by skip
func Caller(skip int) *Frame {
	f := &Frame{
		stack: Pool.Get(),
	}
	runtime.Callers(skip+stackSkip, f.stack.pc)
	return f
}

// Free returns the frame to the stack pool
func (f *Frame) Free() {
	f.stack.Free()
}

// config sets underluting runtime.Frame,
// panics if there is no runtime.Frame
func (f *Frame) config() {
	if f.frame == nil {
		if len(f.stack.pc) > 0 {
			n, _ := runtime.CallersFrames(f.stack.pc).Next()
			if n.PC == 0 {
				panic("cannot configure frame, no such frame")
			}
			f.frame = &n
		}
	}
}

// PC returns the frame PC
func (f *Frame) PC() uintptr {
	return f.stack.pc[0]
}

// Frame returns the runtime.Frame
func (f *Frame) Frame() *runtime.Frame {
	if f.frame == nil {
		f.config()
	}
	return f.frame
}

// Line returns the line number of the frame
func (f *Frame) Line() int {
	if f.line == 0 {
		f.config()
		f.line = f.frame.Line
	}
	return f.line
}

// File returns the file source of the frame
func (f *Frame) File() *src {
	if f.file == nil {
		f.config()
		f.parseFile()
	}
	return f.file
}

// Func returns the func source of the frame
func (f *Frame) Func() *src {
	if f.fn == nil {
		f.config()
		f.parseFuncPkg()
	}
	return f.fn
}

// Pkg returns the package source of the frame
func (f *Frame) Pkg() *src {
	if f.pkg == nil {
		f.config()
		f.parseFuncPkg()
	}
	return f.pkg
}

// frame returns the next caller frame
// from the stack of callers provided
func (f *Frame) Build() *Frame {
	f.config()
	f.line = f.frame.Line
	f.parseFile()
	f.parseFuncPkg()
	return f
}

func (f *Frame) parseFuncPkg() {
	i1 := strings.LastIndex(f.frame.Function, "/") + 1
	i2 := i1 + strings.Index(f.frame.Function[i1:], ".")
	f.parseFunc(i2)
	f.parsePkg(i1, i2)
}

func (f *Frame) parseFunc(i int) {
	f.fn = &src{
		Name: f.frame.Function[i+1:],
		Dir:  f.frame.Function[:i],
		Path: f.frame.Function,
	}
}

func (f *Frame) parsePkg(i1, i2 int) {
	f.pkg = &src{
		Name: f.frame.Function[i1:i2],
		Dir:  f.frame.Function[:i1-1],
		Path: f.frame.Function[:i2],
	}
}

func (f *Frame) parseFile() {
	i := strings.LastIndex(f.frame.File, "/")
	f.file = &src{
		Name: f.frame.File[i+1:],
		Dir:  f.frame.File[:i],
		Path: f.frame.File,
	}
}
