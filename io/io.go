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

package io

import (
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/path"
)

// Functions are the functions available to templates
var Functions = template.FuncMap{
	"inc":  func(i int) int { return i + 1 },
	"add":  func(x, y int) int { return x + y },
	"sub":  func(x, y int) int { return x - y },
	"args": func(args ...any) []any { return args },
	"append": func(s ...string) string {
		b := strings.Builder{}
		for _, v := range s {
			b.WriteString(v)
		}
		return b.String()
	},
}

// IO is a reader/writer that combines the incorporates
// the funcationality of a template, buffer and io.ReadWriter.
type IO struct {
	buf  *buffer.Buffer     // buffer for readWriter
	out  io.ReadWriter      // output readwriter
	tmp  *template.Template // template
	name string             // name of template
}

// New constructs a new IO
func New(output ...io.ReadWriter) *IO {
	return &IO{
		buf: buffer.Pool.Get(),
		out: os.Stdin,
	}
}

// Close frees the buffer and empties the IO
func (i *IO) Close() {
	i.buf.Free()
	i = nil
}

// Reset resets the name, buffer and output writer of the IO
func (i *IO) Reset() *IO {
	i.out = os.Stdin
	i.buf.Reset()
	return i
}

// Copy returns a copy of the IO
func (i *IO) Copy() *IO {
	n := &IO{
		out:  i.out,
		name: i.name,
	}
	n.buf = i.buf.Copy()
	if i.tmp != nil {
		n.tmp = template.Must(n.tmp.Clone())
	}
	return n
}

// --------------------------------------------------------------------------- /
// IO Output getters and setters
// --------------------------------------------------------------------------- /

// Writer sets the IO ouput writers
func (i *IO) SetOut(out ...io.ReadWriter) *IO {
	if out != nil {
		if len(out) == 1 {
			i.out = out[0]
			return i
		}
		i.out = MultiReadWriter(out...)
	}
	return i
}

// AddOut adds output writers to the IO
func (i *IO) AddOut(out ...io.ReadWriter) *IO {
	if out != nil {
		if i.out != nil {
			i.out = AppendReadWriter(i.out, out...)
			return i
		}
		i.out = MultiReadWriter(out...)
	}
	return i
}

// Out returns the output writer of the IO
func (i *IO) Out() io.Writer {
	return i.out
}

// OutFiles returns all files in the output writer of the IO
// if the output writer contains files, otherwise nil
func (i *IO) OutFiles() []*os.File {
	return ReadWriterFiles(i.out)
}

// --------------------------------------------------------------------------- /
// IO Buffer getters and setters
// --------------------------------------------------------------------------- /

// Buffer returns the buffer of the IO
func (i *IO) Buffer() *buffer.Buffer {
	return i.buf
}

// Prepend prepends b to the buffer
func (i *IO) Prepend(b []byte) (int, error) {
	return i.buf.Prepend(b)
}

// Append appends b to the buffer
func (i *IO) Append(b []byte) (int, error) {
	return i.buf.Write(b)
}

// AppendByte appends b to the buffer
func (i *IO) AppendByte(b byte) error {
	return i.buf.WriteByte(b)
}

// AppendBool appends b to the buffer
func (i *IO) AppendBool(b bool) (int, error) {
	return i.buf.WriteBool(b)
}

// AppendInt appends b to the buffer
func (i *IO) AppendInt(b int) (int, error) {
	return i.buf.WriteInt(b)
}

// AppendUint appends b to the buffer
func (i *IO) AppendUint(b uint) (int, error) {
	return i.buf.WriteUint(b)
}

// AppendFloat appends b to the buffer
func (i *IO) AppendFloat(b float64) (int, error) {
	return i.buf.WriteFloat(b)
}

// AppendString appends s to the buffer
func (i *IO) AppendString(s string) (int, error) {
	return i.buf.WriteString(s)
}

// AppendStrings appends s to the buffer
func (i *IO) AppendStrings(s ...string) (int, error) {
	return i.buf.WriteStrings(s...)
}

// Execute executes the IO template with data
// and appends the result to the buffer
func (i *IO) Execute(data any) error {
	return i.tmp.Execute(i.buf, data)
}

// Inject treates the buffer as a template and inserts data into the template values,
// synonymous with combining ParseBuffer and Execute
func (i *IO) Inject(data any) error {
	err := i.Generate()
	if err != nil {
		return err
	}
	return i.Execute(data)
}

// ReadFile reads the file at path into the buffer,
// appending to any previous data,
// and sets the name of the IO to the file name
func (i *IO) ReadFile(path *path.Path) (int, error) {
	b, err := os.ReadFile(path.Path())
	if err != nil {
		return 0, err
	}
	l, err := i.Append(b)
	if err != nil {
		return l, err
	}
	i.name = path.File().Name
	return l, nil
}

// ExecTmplFile parses the file at path as a template,
// sets the template of the IO  and executes the template with data,
// overriding any previous template or buffered data
func (i *IO) ExecTmplFile(path *path.Path, data any) (err error) {
	_, err = i.ReadFile(path)
	if err != nil {
		return
	}
	err = i.Generate()
	if err != nil {
		return
	}
	err = i.Execute(data)
	return
}

// MustPrepend prepends b to the buffer, panics if an error occurs
func (i *IO) MustPrepend(b []byte) *IO {
	MustIO(i.Prepend(b))
	return i
}

// MustAppend appends b to the buffer, panics if an error occurs
func (i *IO) MustAppend(b []byte) *IO {
	MustIO(i.Append(b))
	return i
}

// MustAppendByte appends b to the buffer, panics if an error occurs
func (i *IO) MustAppendByte(b byte) *IO {
	Must(i.AppendByte(b))
	return i
}

// MustAppendBool appends b to the buffer, panics if an error occurs
func (i *IO) MustAppendBool(b bool) *IO {
	MustIO(i.AppendBool(b))
	return i
}

// MustAppendInt appends b to the buffer, panics if an error occurs
func (i *IO) MustAppendInt(b int) *IO {
	MustIO(i.AppendInt(b))
	return i
}

// MustAppendUint appends b to the buffer, panics if an error occurs
func (i *IO) MustAppendUint(b uint) *IO {
	MustIO(i.AppendUint(b))
	return i
}

// MustAppendFloat appends b to the buffer, panics if an error occurs
func (i *IO) MustAppendFloat(b float64) *IO {
	MustIO(i.AppendFloat(b))
	return i
}

// MustAppendString appends s to the buffer, panics if an error occurs
func (i *IO) MustAppendString(s string) *IO {
	MustIO(i.AppendString(s))
	return i
}

// MustAppendStrings appends s to the buffer, panics if an error occurs
func (i *IO) MustAppendStrings(s ...string) *IO {
	MustIO(i.AppendStrings(s...))
	return i
}

// MustExecute executes the IO template with data
// and appends the result to the buffer,
// panics if an error occurs
func (i *IO) MustExecute(data any) *IO {
	Must(i.Execute(data))
	return i
}

// Inject treates the buffer as a template and inserts data into the template values,
// synonymous with combining ParseBuffer and Execute, panics if an error occurs
func (i *IO) MustInject(data any) *IO {
	Must(i.Inject(data))
	return i
}

// MustReadFile reads the file at path into the buffer,
// appending to any previous data,
// and sets the name of the IO to the file name,
// panics if an error occurs
func (i *IO) MustReadFile(path *path.Path) *IO {
	MustIO(i.ReadFile(path))
	return i
}

// MustExecTmplFile parses the file at path as a template,
// sets the template of the IO  and executes the template with data,
// overriding any previous template or buffered data, panics if an error occurs
func (i *IO) MustExecTmplFile(path *path.Path, data any) *IO {
	Must(i.ExecTmplFile(path, data))
	return i
}

// --------------------------------------------------------------------------- /
// IO Template getters and setters
// --------------------------------------------------------------------------- /

// SetTemplate sets the template of the IO
func (i *IO) SetTemplate(t *template.Template) *IO {
	i.name = t.Name()
	i.tmp = t
	return i
}

// Parse parses the template string and sets the template of the IO
func (i *IO) Parse(name, s string) (err error) {
	i.tmp, err = template.New(name).Funcs(Functions).Parse(s)
	if err != nil {
		return
	}
	i.name = name
	return nil
}

// Generate moves the buffer to the template of the IO
func (i *IO) Generate() (err error) {
	name := i.name
	if name == "" {
		if f, ok := i.out.(*os.File); ok {
			name = path.New(f.Name()).File().Name
		} else {
			name = "io"
		}
	}
	i.tmp, err = template.New(name).Funcs(Functions).Parse(i.buf.String())
	if err != nil {
		return
	}
	i.name = name
	i.buf.Reset()
	return
}

// Template returns the template of the IO
func (i *IO) Template() *template.Template {
	return i.tmp
}

// --------------------------------------------------------------------------- /
// IO Reader methods
// --------------------------------------------------------------------------- /

// Read reads from the IO's input into the given bytes.
func (i *IO) Read(bytes []byte) (int, error) {
	return i.out.Read(bytes)
}

// ReadBuffer reads from the IO's buffer into the given bytes.
func (i *IO) ReadBuffer(bytes []byte) (int, error) {
	return i.buf.Read(bytes)
}

// MustRead reads from the IO's input into the given bytes
// and panics if an error occurs.
func (i *IO) MustRead(bytes []byte) *IO {
	MustIO(i.Read(bytes))
	return i
}

// MustReadBuffer reads from the IO's buffer into the given bytes
// and panics if an error occurs.
func (i *IO) MustReadBuffer(bytes []byte) *IO {
	MustIO(i.ReadBuffer(bytes))
	return i
}

// --------------------------------------------------------------------------- /
// IO Writer methods
// --------------------------------------------------------------------------- /

// Output writes the buffer to the output writer
func (i *IO) Output() (int, error) {
	defer i.buf.Reset()
	return i.out.Write(i.buf.Bytes())
}

// Write writes b to the buffer and the buffer to the output writer
func (i *IO) Write(b []byte) (int, error) {
	_, err := i.buf.Write(b)
	if err != nil {
		return 0, err
	}
	return i.Output()
}

// MustOutput writes the buffer to the output writer, panics if an error occurs
func (i *IO) MustOutput() *IO {
	_, err := i.Output()
	Must(err)
	return i
}

// MustWrite writes b to the buffer and the buffer
// to the output writer, panics if an error occurs
func (i *IO) MustWrite(b []byte) *IO {
	_, err := i.Write(b)
	Must(err)
	return i
}

// --------------------------------------------------------------------------- /
// Helps
// --------------------------------------------------------------------------- /

// Must panics if err is not nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustWrite panics if err is not nil
func MustIO(_ int, err error) {
	if err != nil {
		panic(err)
	}
}
