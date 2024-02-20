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

package test

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// ----------------------------------------------------------------------------
// TEST

type Test struct {
	*sync.Mutex
	*Config
	cnt int
}

type Config struct {
	t       *testing.T
	log     *Buffer
	Require bool
	Print   bool
	Trace   bool
	Detail  bool
	Msg     string
}

func New(t *testing.T, config ...*Config) (T *Test) {
	var c *Config
	if len(config) > 0 {
		c = config[0]
		c.t = t
	} else {
		c = &Config{
			t:      t,
			Trace:  true,
			Detail: true,
			Msg:    "%s",
		}
	}
	T = &Test{&sync.Mutex{}, c, 0}
	t.Cleanup(T.exit)
	return
}

func (t *Test) Equal(expected any, actual any, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("Equal", reflect.DeepEqual(expected, actual), expected, actual, msgArgs)
}

func (t *Test) NotEqual(expected any, actual any, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("NotEqual", !reflect.DeepEqual(expected, actual), expected, actual, msgArgs)
}

func (t *Test) True(actual bool, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("True", actual, true, actual, msgArgs)
}

func (t *Test) False(actual bool, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("False", !actual, false, actual, msgArgs)
}

func (t *Test) Nil(actual any, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("Nil", actual == nil, nil, actual, msgArgs)
}

func (t *Test) NotNil(actual any, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("NotNil", actual != nil, nil, actual, msgArgs)
}

func (t *Test) Error(err error, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("Error", err != nil, nil, err, msgArgs)
}

func (t *Test) NoError(err error, msgArgs ...any) bool {
	t.t.Helper()
	return t.run("NoError", err == nil, nil, err, msgArgs)
}

func (t *Test) Fail(msgArgs ...any) {
	t.t.Helper()
	t.run("Fail", false, nil, nil, msgArgs)
}

func (t *Test) Pass(msgArgs ...any) {
	t.t.Helper()
	t.run("Pass", true, nil, nil, msgArgs)
}

func (t *Test) run(test string, pass bool, expected any, actual any, msgArgs []any) bool {
	t.t.Helper()
	t.Lock()
	defer t.Unlock()
	var msg string
	if len(msgArgs) > 0 {
		msg = fmt.Sprintf(t.Msg, msgArgs...)
	} else if t.Msg != "%s" {
		msg = t.Msg
	}
	m := Msg(t.cnt, t.t.Name()+"."+test, pass, expected, actual, true, true, msg)
	if !pass {
		if t.log == nil {
			t.log = NewBuffer()
		}
		t.log.Write(m.Bytes())
		if t.Require {
			t.exit()
		}
	} else if t.Print {
		os.Stdout.Write(m.Bytes())
	}
	t.cnt++
	return pass
}

func (t *Test) exit() {
	if t.log != nil {
		os.Stdout.WriteString("--- FAIL: " + t.t.Name() + "\n")
		os.Stdout.Write(t.log.Bytes())
		os.Exit(1)
	}
}

// ----------------------------------------------------------------------------
// LIBRARY

func run(t *testing.T, cnt int, print, require bool, actual, expected any, msg string) (pass bool) {
	t.Helper()
	pass = actual == expected
	m := Msg(cnt, t.Name(), pass, expected, actual, true, true, msg)
	if !pass {
		os.Stdout.WriteString("--- FAIL: " + t.Name() + "\n")
		if require {
			t.Fatalf(m.String())
		}
		t.Errorf(m.String())
	} else if print {
		os.Stdout.Write(m.Bytes())
	}
	return
}

func Print(t *testing.T, actual, expected any, msg ...any) (pass bool) {
	t.Helper()
	return run(t, -1, true, false, actual, expected, userMsg(msg...))
}

func Printr(t *testing.T, actual, expected any, msg ...any) (pass bool) {
	t.Helper()
	return run(t, -1, true, true, actual, expected, userMsg(msg...))
}

func Assert(t *testing.T, actual, expected any, msg ...any) (pass bool) {
	t.Helper()
	return run(t, -1, false, false, actual, expected, userMsg(msg...))
}

func Require(t *testing.T, actual, expected any, msg ...any) (pass bool) {
	t.Helper()
	return run(t, -1, false, true, actual, expected, userMsg(msg...))
}

func userMsg(msg ...any) (s string) {
	if len(msg) > 0 {
		s = msg[0].(string)
	}
	if len(msg) > 1 {
		s = fmt.Sprintf(s, msg[1:]...)
	}
	return
}

func Msg(num int, test string, pass bool, expected any, actual any, trace, detail bool, msg string) (m *Buffer) {
	m = NewBuffer()
	m.WriteByte('\n')
	if pass {
		m.WriteString("PASS:     ")
	} else {
		m.WriteString("FAIL:     ")
	}
	if num > -1 {
		m.WriteByte('#')
		m.WriteString(strconv.Itoa(num))
		m.WriteByte(' ')
	}
	if test != "" {
		m.WriteString(test)
	}
	if msg != "" {
		m.WriteString(": ")
		m.WriteString(msg)
	}
	m.WriteByte('\n')
	if trace {
		m.WriteString("src:      ")
		m.Write(Trace(3).Bytes())
		m.WriteByte('\n')
	}
	if detail {
		format := "%#[1]v\n"
		m.WriteString("expected: ")
		m.WriteString(fmt.Sprintf(format, expected))
		m.WriteString("actual:   ")
		m.WriteString(fmt.Sprintf(format, actual))
		m.WriteByte('\n')
	}
	return
}

func Trace(skip int) (t *Buffer) {
	pc := make([]uintptr, 1)
	runtime.Callers(skip+1, pc)
	f, _ := runtime.CallersFrames(pc).Next()
	t = NewBuffer()
	if f.PC != 0 {
		// packakge name
		pkgStart := strings.LastIndex(f.Function, `/`) + 1
		pkgEnd := strings.Index(f.Function[pkgStart:], `.`) + pkgStart
		t.WriteString(f.Function[:pkgEnd])
		t.WriteString(`.`)
		// file name
		fileStart := strings.LastIndex(f.File, `/`) + 1
		fileEnd := strings.LastIndex(f.File, `.`)
		t.WriteString(f.File[fileStart:fileEnd])
		// file line
		t.WriteString(` line `)
		t.WriteString(strconv.Itoa(f.Line))
		return
	}
	t.WriteString(`unknown source`)
	return
}

func PrintTable(data [][]string, header bool) {
	var (
		colDel  = " | "
		rowDel  = "\n"
		hColDel = "-+-"
		hRowSpc = "-"
		Space   = " "
		t       = ""
	)
	size := make([]int, len(data[0]))
	for i := range data {
		for j := range data[i] {
			if len(data[i][j]) > size[j] {
				size[j] = len(data[i][j])
			}
		}
	}
	for i := range data {
		if i == 1 && header {
			for j := range data[i] {
				if j > 0 {
					t += hColDel
				}
				t += strings.Repeat(hRowSpc, size[j])
			}
			t += rowDel
		}
		for j := range data[i] {
			if j > 0 {
				t += colDel
			}
			t += data[i][j] + strings.Repeat(Space, size[j]-len(data[i][j]))
		}
		t += rowDel
	}
	fmt.Print(t)
}

// ----------------------------------------------------------------------------
// Buffer

type Buffer []byte

func NewBuffer() *Buffer {
	b := make(Buffer, 0, 256)
	return &b
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	*b = append(*b, s...)
	return len(s), nil
}

func (b *Buffer) WriteByte(c byte) error {
	*b = append(*b, c)
	return nil
}

func (b *Buffer) String() string {
	return string(*b)
}

func (b *Buffer) Bytes() []byte {
	return *b
}

func (b *Buffer) Reset() {
	*b = (*b)[:0]
}
