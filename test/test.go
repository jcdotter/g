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
	"log"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
	suite "github.com/stretchr/testify/suite"
)

const (
	VERSION = "0.0.3"
)

var (
	Assert  = &assert.Assertions{}
	Require = &require.Assertions{}
	Suite   = &suite.Suite{}
	Mock    = &mock.Mock{}
	delim   = `.`
)

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

func (t *Test) output(test string, pass bool, expected any, actual any, msgArgs []any) {
	t.Lock()
	msg := ""
	if t.willPrint {
		if t.PrintTest || (t.PrintFail && !pass) {
			msg = t.buildMsg(test, pass, expected, actual, msgArgs)
			log.Print(msg)
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

func (t *Test) buildMsg(test string, pass bool, expected any, actual any, msgArgs []any) string {
	msg := "\n#" + strconv.Itoa(t.cnt) + " test '" + test + "' "
	if pass {
		msg += "succeeded: "
	} else {
		msg += "failed: "
	}
	if t.Msg != "" {
		msg += fmt.Sprintf(t.Msg+"\n", msgArgs...)
	}
	if t.PrintTrace {
		msg += "  src:\t\t" + trace(3) + "\n"
	}
	if t.PrintDetail {
		format := "\t%#[1]v\n"
		msg += "  expected:" + fmt.Sprintf(format, expected)
		msg += "  actual:" + fmt.Sprintf(format, actual)
	}
	return msg
}

func trace(i int) string {
	pc, fl, ln, ok := runtime.Caller(int(i + 1))
	if ok {
		fs := strings.Split(fl, `/`)
		gf := strings.Split(fs[len(fs)-1], delim)[0]
		fn := strings.Split(runtime.FuncForPC(pc).Name(), delim)
		pt := strings.Replace(fn[0], `/`, delim, -1)
		s := []string{pt, gf}
		s = append(s, fn[1:]...)
		return strings.Join(s, delim) + ` line ` + fmt.Sprint(ln)
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
