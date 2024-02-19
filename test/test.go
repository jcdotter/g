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
	t           *testing.T
	FailFatal   bool
	PrintTest   bool
	PrintFail   bool
	PrintTrace  bool
	PrintDetail bool
	Truncate    bool
	Msg         string
	willPrint   bool
}

func New(t *testing.T, config *Config) *Test {
	config.t = t
	config.willPrint = config.PrintTest || config.PrintFail || config.PrintTrace || config.PrintDetail
	return &Test{&sync.Mutex{}, config, 0}
}

func (t *Test) Equal(expected any, actual any, msgArgs ...any) bool {
	pass := reflect.DeepEqual(expected, actual)
	t.output("Equal", pass, expected, actual, msgArgs)
	return pass
}

func (t *Test) NotEqual(expected any, actual any, msgArgs ...any) bool {
	pass := !reflect.DeepEqual(expected, actual)
	t.output("NotEqual", pass, expected, actual, msgArgs)
	return pass
}

func (t *Test) True(actual bool, msgArgs ...any) bool {
	pass := actual
	t.output("True", pass, true, actual, msgArgs)
	return pass
}

func (t *Test) False(actual bool, msgArgs ...any) bool {
	pass := !actual
	t.output("False", pass, false, actual, msgArgs)
	return pass
}

func (t *Test) Nil(actual any, msgArgs ...any) bool {
	pass := actual == nil
	t.output("Nil", pass, nil, actual, msgArgs)
	return pass
}

func (t *Test) NotNil(actual any, msgArgs ...any) bool {
	pass := actual != nil
	t.output("NotNil", pass, nil, actual, msgArgs)
	return pass
}

func (t *Test) Error(err error, msgArgs ...any) bool {
	pass := err != nil
	t.output("Error", pass, nil, err, msgArgs)
	return pass
}

func (t *Test) NoError(err error, msgArgs ...any) bool {
	pass := err == nil
	t.output("NoError", pass, nil, err, msgArgs)
	return pass
}

func (t *Test) Fail(msgArgs ...any) {
	t.output("Fail", false, nil, nil, msgArgs)
}

func (t *Test) Pass(msgArgs ...any) {
	t.output("Pass", true, nil, nil, msgArgs)
}

func (t *Test) output(test string, pass bool, expected any, actual any, msgArgs []any) {
	t.Lock()
	msg := ""
	if t.willPrint {
		if t.PrintTest || (t.PrintFail && !pass) {
			t.t.Log(Msg(t.cnt, test, pass, expected, actual, t.PrintTrace, t.PrintDetail, t.Msg, msgArgs...))
		}
	}
	t.cnt++
	if !pass {
		if t.FailFatal {
			t.Unlock()
			t.t.FailNow()
		}
		t.t.Log(msg)
		defer t.t.Fail()
	}
	t.Unlock()
}

// ----------------------------------------------------------------------------
// LIBRARY

func Assert(t *testing.T, actual, expected any, msg ...any) {
	if actual != expected {
		t.Errorf(Msg(-1, "", false, expected, actual, true, true, userMsg(msg...)))
	} else {
		t.Logf(Msg(-1, "", true, expected, actual, true, true, userMsg(msg...)))
	}
}

func Require(t *testing.T, actual, expected any, msg ...any) {
	if actual != expected {
		t.Fatalf(Msg(-1, "", false, expected, actual, true, true, userMsg(msg...)))
	} else {
		t.Logf(Msg(-1, "", true, expected, actual, true, true, userMsg(msg...)))
	}
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

func Msg(num int, test string, pass bool, expected any, actual any, trace, detail bool, msg string, args ...any) string {
	m := strings.Builder{}
	m.Grow(256)
	m.WriteByte('\n')
	if num > 0 {
		m.WriteString("#")
		m.WriteString(strconv.Itoa(num))
		m.WriteByte(' ')
	}
	if pass {
		m.WriteString("PASS: ")
	} else {
		m.WriteString("FAIL: ")
	}
	if test != "" {
		m.WriteString("test '")
		m.WriteString(test)
		m.WriteString("'. ")
	}
	if msg != "" {
		m.WriteString(fmt.Sprintf(msg, args...))
	}
	m.WriteByte('\n')
	if trace {
		m.WriteString("  src:\t\t")
		m.WriteString(Trace(3))
		m.WriteByte('\n')
	}
	if detail {
		format := "\t%#[1]v\n"
		m.WriteString("  expected:")
		m.WriteString(fmt.Sprintf(format, expected))
		m.WriteString("  actual:")
		m.WriteString(fmt.Sprintf(format, actual))
		m.WriteByte('\n')
	}
	return m.String()
}

func Trace(skip int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(skip+1, pc)
	f, _ := runtime.CallersFrames(pc).Next()
	if f.PC != 0 {
		s := strings.Builder{}
		s.Grow(64)
		// packakge name
		pkgStart := strings.LastIndex(f.Function, `/`) + 1
		pkgEnd := strings.Index(f.Function[pkgStart:], `.`) + pkgStart
		s.WriteString(f.Function[pkgStart:pkgEnd])
		s.WriteString(`.`)
		// file name
		fileStart := strings.LastIndex(f.File, `/`) + 1
		fileEnd := strings.LastIndex(f.File, `.`)
		s.WriteString(f.File[fileStart:fileEnd])
		// file line
		s.WriteString(` line `)
		s.WriteString(strconv.Itoa(f.Line))
		return s.String()
	}
	return `unknown.source`
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
