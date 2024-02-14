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

package logger

import (
	"strconv"
	"testing"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/stack"
	"github.com/jcdotter/go/test"
)

var (
	testLogger *Logger
)

func init() {
	testLogger = New().Writers(buffer.New()).Build()
}

var config = &test.Config{
	//PrintTest:   true,
	PrintFail:   true,
	PrintTrace:  true,
	PrintDetail: true,
	FailFatal:   true,
	Msg:         "%s",
}

func TestAll(t *testing.T) {
	TestLevel(t)
	TestTime(t)
	TestService(t)
	TestCallId(t)
	TestFrame(t)
	TestCaller(t)
	TestStatic(t)
	TestFields(t)
	TestKeyVals(t)
	TestMessage(t)
}

func TestLevel(t *testing.T) {
	l := New().LevelKey("lvl").Build()
	gt := test.New(t, config)
	gt.Equal(`{"lvl":"debug"`, l.level(LevelDebug).String())
	gt.Equal(`{"lvl":"info"`, l.level(LevelInfo).String())
	gt.Equal(`{"lvl":"warn"`, l.level(LevelWarn).String())
	gt.Equal(`{"lvl":"error"`, l.level(LevelError).String())
	gt.Equal(`{"lvl":"fatal"`, l.level(LevelFatal).String())
	gt.Equal(`{"lvl":"panic"`, l.level(LevelPanic).String())
}

func TestTime(t *testing.T) {
	f := "2006-01-02 15:04:05.000"
	l := New().TimeKey("time").TimeFmt(f).Build()
	gt := test.New(t, config)
	gt.Equal(`,"time":"`+l.clock.time.Format(f)+`"`, l.time().String())
}

func TestService(t *testing.T) {
	l := New().ServiceKey("service").ServiceName("test").Build()
	gt := test.New(t, config)
	gt.Equal(`,"service":"test"`, l.service().String())
}

func TestCallId(t *testing.T) {
	l := New().CallIdKey("callid").Build()
	gt := test.New(t, config)
	gt.Equal(`,"callid":"123"`, l.callid("123").String())
}

func TestFrame(t *testing.T) {
	f := stack.Caller(0).Build()
	gt := test.New(t, config)
	gt.Equal(`github.com/jcdotter/grpg/utils/logger`, f.Pkg().Path)
	gt.Equal(`github.com/jcdotter/grpg/utils`, f.Pkg().Dir)
	gt.Equal(`logger`, f.Pkg().Name)
	gt.Equal(`logger_test.go`, f.File().Name)
	gt.Equal(`TestFrame`, f.Func().Name)
	gt.NotEqual(0, f.Line())
}

func TestCaller(t *testing.T) {
	l := New().CallerKey("caller").Build()
	gt := test.New(t, config)
	gt.Equal(`,"pkg":"github.com/jcdotter/grpg/utils/logger","caller":"logger_test.go:109","fn":"TestCaller"`, l.encCaller(1).String())
}

func TestStatic(t *testing.T) {
	l := New().
		AddStaticField("key", "value").
		AddStaticField("key2", "value2").Build()
	gt := test.New(t, config)
	gt.Equal(`,"key":"value","key2":"value2"`, l.staticFields().String())
}

func TestFields(t *testing.T) {
	l := New().AddField("key", func(l *Logger) any {
		return "value"
	}).AddField("key2", func(l *Logger) any {
		return "value2"
	}).Build()
	gt := test.New(t, config)
	gt.Equal(`,"key":"value","key2":"value2"`, l.fields().String())
}

func TestKeyVals(t *testing.T) {
	l := New().Build()
	gt := test.New(t, config)
	gt.Equal(`,"key":"value","key2":"value2"`, l.keyVals("key", "value", "key2", "value2").String())
}

func TestMessage(t *testing.T) {
	l := New().MessageKey("msg").Build()
	gt := test.New(t, config)
	gt.Equal(`,"msg":"test"`, l.message("test").String())
}

func BenchmarkLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testLogger.Write(LevelInfo, "test"+strconv.Itoa(i))
	}
}
