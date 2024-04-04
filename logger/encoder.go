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
	"encoding/json"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/stack"
)

// default encoder settings
var defaultEncoder *Encoder = &Encoder{
	LevelKey:    "level",
	TimeKey:     "ts",
	ServiceKey:  "svc",
	CallIdKey:   "cid",
	PackageKey:  "pkg",
	CallerKey:   "src",
	FunctionKey: "fn",
	MessageKey:  "msg",
	LogStart:    '{',
	LogEnd:      '}',
	LogSep:      '\n',
	ElemSep:     ',',
	KeyValSep:   ':',
	Quote:       '"',
	TimeFmt:     "2006-01-02 15:04:05.000",
}

var defBufSize = 24

type Encoder struct {
	LevelKey       string
	TimeKey        string
	ServiceKey     string
	ServiceName    string
	CallIdKey      string
	PackageKey     string
	CallerKey      string
	FunctionKey    string
	MessageKey     string
	LevelIndex     []uint8
	LevelKeyBuffer *buffer.Buffer // eg. `{"level":` (every log entry must have a level)
	LevelBuffer    *buffer.Buffer // eg. `{"level":"info"`
	TimeKeyBuffer  *buffer.Buffer // eg. `,"ts":`
	TimeBuffer     *buffer.Buffer // eg. `,"ts":"2006-01-02 15:04:05.000"`
	ServiceBuffer  *buffer.Buffer // eg. `,"svc":"my-service"` (service name doesn't change)
	CallIdBuffer   *buffer.Buffer // eg. `,"callid":`
	PackageBuffer  *buffer.Buffer // eg. `,"pkg":`
	CallerBuffer   *buffer.Buffer // eg. `,"src":`
	FunctionBuffer *buffer.Buffer // eg. `,"fn":`
	StaticBuffer   *buffer.Buffer // eg. `,"key":"value"`
	MessageBuffer  *buffer.Buffer // eg. `,"msg":`
	EndBuffer      []byte         // eg. `}`
	LogStart       byte           // eg. `{`
	LogEnd         byte           // eg. `}`
	LogSep         byte           // eg. `\n`
	ElemSep        byte           // eg. `,`
	KeyValSep      byte           // eg. `:`
	Quote          byte           // eg. `"`
	TimeFmt        string
	CallerCache    map[uintptr][]byte
}

func NewEncoder() (e *Encoder) {
	return &Encoder{
		LevelKey:    defaultEncoder.LevelKey,
		TimeKey:     defaultEncoder.TimeKey,
		ServiceKey:  defaultEncoder.ServiceKey,
		CallIdKey:   defaultEncoder.CallIdKey,
		PackageKey:  defaultEncoder.PackageKey,
		CallerKey:   defaultEncoder.CallerKey,
		FunctionKey: defaultEncoder.FunctionKey,
		MessageKey:  defaultEncoder.MessageKey,
		/* LevelKeyBuffer: buffer.Make(defBufSize),
		LevelBuffer:    buffer.Make(256), */
		/* TimeKeyBuffer:  buffer.Make(defBufSize),
		TimeBuffer:     buffer.Make(64),
		ServiceBuffer:  buffer.Make(defBufSize),
		CallIdBuffer:   buffer.Make(defBufSize),
		PackageBuffer:  buffer.Make(defBufSize),
		CallerBuffer:   buffer.Make(defBufSize),
		FunctionBuffer: buffer.Make(defBufSize),
		StaticBuffer:   buffer.Make(defBufSize), */
		//MessageBuffer: buffer.Make(defBufSize),
		EndBuffer: []byte{defaultEncoder.LogEnd, defaultEncoder.LogSep},
		LogStart:  defaultEncoder.LogStart,
		LogEnd:    defaultEncoder.LogEnd,
		LogSep:    defaultEncoder.LogSep,
		ElemSep:   defaultEncoder.ElemSep,
		KeyValSep: defaultEncoder.KeyValSep,
		Quote:     defaultEncoder.Quote,
		TimeFmt:   defaultEncoder.TimeFmt,
	}
}

// PresetBuffers pre-sets the encoder buffers
func (e *Encoder) PresetBuffers(c *Config) {
	e.PresetBuffer(true, &e.LevelKeyBuffer, defBufSize, e.LogStart, e.LevelKey)
	e.PresetBuffer(c.LogTime, &e.TimeKeyBuffer, defBufSize, e.ElemSep, e.TimeKey)
	e.PresetBuffer(c.LogTime, &e.TimeBuffer, 64, e.ElemSep, e.TimeKey)
	e.PresetBuffer(c.LogService, &e.ServiceBuffer, defBufSize, e.ElemSep, e.ServiceKey)
	e.PresetBuffer(true, &e.CallIdBuffer, defBufSize, e.ElemSep, e.CallIdKey)
	e.PresetBuffer(c.LogPackage, &e.PackageBuffer, defBufSize, e.ElemSep, e.PackageKey)
	e.PresetBuffer(c.LogCaller, &e.CallerBuffer, defBufSize, e.ElemSep, e.CallerKey)
	e.PresetBuffer(c.LogFunction, &e.FunctionBuffer, defBufSize, e.ElemSep, e.FunctionKey)
	e.PresetBuffer(true, &e.MessageBuffer, defBufSize, e.ElemSep, e.MessageKey)

	if c.LogService {
		e.BufferString(e.ServiceBuffer, e.ServiceName)
	}
	e.BufferLevels()
}

func (e *Encoder) PresetBuffer(use bool, b **buffer.Buffer, size int, sep byte, key string) {
	if use {
		if *b == nil {
			*b = buffer.Make(size)
		} else {
			(*b).Reset()
		}
		e.BufferKey(*b, sep, key)
	}
}

// BufferKey writes the provided key to the provided buffer
// prepended with the provided separator
func (e *Encoder) BufferKey(b *buffer.Buffer, sep byte, key string) {
	b.WriteByte(sep)
	b.WriteByte(e.Quote)
	b.WriteString(key)
	b.WriteByte(e.Quote)
	b.WriteByte(e.KeyValSep)
}

// BufferVal writes the provided value to the provided buffer
func (e *Encoder) BufferVal(b *buffer.Buffer, val any) {
	j, _ := json.Marshal(val)
	b.Write(j)
}

// BufferString writes the provided non-literal string to
// the provided buffer as a literal string
func (e *Encoder) BufferString(b *buffer.Buffer, s string) {
	b.WriteByte(e.Quote)
	b.WriteString(s)
	b.WriteByte(e.Quote)
}

// BufferBytes writes the provided bytes to the provided buffer
func (e *Encoder) BufferBytes(b *buffer.Buffer, s []byte) {
	b.WriteByte(e.Quote)
	b.Write(s)
	b.WriteByte(e.Quote)
}

// CreateKeyVal writes the provided key and value to the provided buffer
func (e *Encoder) BufferKeyVal(b *buffer.Buffer, key string, val any) {
	e.BufferKey(b, e.ElemSep, key)
	e.BufferVal(b, val)
}

// BufferKeyValString writes the provided key and value to the provided buffer
func (e *Encoder) BufferKeyValString(b *buffer.Buffer, key string, val string) {
	e.BufferKey(b, e.ElemSep, key)
	e.BufferString(b, val)
}

func (e *Encoder) BufferLevels() {
	e.LevelBuffer = buffer.Make(256)
	e.LevelIndex = make([]uint8, len(levelNameIndex))
	for i := 0; i < len(levelNameIndex)-1; i++ {
		e.LevelBuffer.WriteBytes(e.LevelKeyBuffer.Bytes())
		e.LevelBuffer.WriteByte(e.Quote)
		e.LevelBuffer.WriteString(Level(i).String())
		e.LevelBuffer.WriteByte(e.Quote)
		e.LevelIndex[i+1] = uint8(e.LevelBuffer.Len())
	}
}

//------------------------------------------------------------
// Logger encoder elements
//------------------------------------------------------------

type field struct {
	name string
	pre  []byte
	fn   func(*Logger) any
}

type encoding []byte

func (e encoding) String() string {
	return string(e)
}

//------------------------------------------------------------
// Logger element encoders
// encode the logger elements into a byte slice
//------------------------------------------------------------

func (l *Logger) level(ll Level) encoding {
	return l.encoder.LevelBuffer.Bytes()[l.encoder.LevelIndex[ll]:l.encoder.LevelIndex[ll+1]]
}

func (l *Logger) time() encoding {
	if l.config.LogTime {
		if l.clock.Refresh() {
			l.encoder.TimeBuffer.Reset()
			l.encoder.TimeBuffer.Write(l.encoder.TimeKeyBuffer.Bytes())
			l.encoder.BufferBytes(l.encoder.TimeBuffer, l.clock.Cache())
		}
		return l.encoder.TimeBuffer.Bytes()
	}
	return nil
}

func (l *Logger) service() encoding {
	if l.config.LogService {
		return l.encoder.ServiceBuffer.Bytes()
	}
	return nil
}

func (l *Logger) callid(cid string) encoding {
	if len(cid) > 0 {
		l.encoder.BufferString(l.encoder.CallIdBuffer, cid)
		return l.encoder.CallIdBuffer.Bytes()
	}
	return nil
}

func (l *Logger) encCaller(skip int) (enc encoding) {
	if l.config.LogCaller {
		caller := stack.Caller(skip)
		defer caller.Free()
		l.Lock()
		defer l.Unlock()
		var ok bool
		if enc, ok = l.encoder.CallerCache[caller.PC()]; !ok {
			b := buffer.Pool.Get()
			defer b.Free()
			if l.config.LogPackage {
				b.WriteBytes(l.encoder.PackageBuffer.Bytes())
				l.encoder.BufferString(b, caller.Pkg().Path)
			}
			if l.config.LogCaller {
				b.WriteBytes(l.encoder.CallerBuffer.Bytes())
				b.WriteByte(l.encoder.Quote)
				b.WriteString(caller.File().Name)
				b.WriteByte(':')
				b.WriteInt(caller.Line())
				b.WriteByte(l.encoder.Quote)
			}
			if l.config.LogFunction {
				b.WriteBytes(l.encoder.FunctionBuffer.Bytes())
				l.encoder.BufferString(b, caller.Func().Name)
			}
			enc = b.Bytes()
			if l.encoder.CallerCache == nil {
				l.encoder.CallerCache = map[uintptr][]byte{caller.PC(): enc}
			} else {
				l.encoder.CallerCache[caller.PC()] = enc
			}
		}
	}
	return
}

func (l *Logger) staticFields() encoding {
	if l.config.LogStatics {
		return l.encoder.StaticBuffer.Bytes()
	}
	return nil
}

func (l *Logger) fields() encoding {
	if len(l.config.Fields) > 0 {
		b := buffer.Pool.Get()
		defer b.Free()
		for _, f := range l.config.Fields {
			l.encoder.BufferKey(b, l.encoder.ElemSep, string(f.pre))
			l.encoder.BufferVal(b, f.fn(l))
		}
		return b.Bytes()
	}
	return nil
}

func (l *Logger) keyVals(keyvals ...any) encoding {
	if len(keyvals) > 0 {
		b := buffer.Pool.Get()
		defer b.Free()
		for i := 0; i < len(keyvals); i += 2 {
			l.encoder.BufferKey(b, l.encoder.ElemSep, keyvals[i].(string))
			l.encoder.BufferVal(b, keyvals[i+1])
		}
		return b.Bytes()
	}
	return nil
}

func (l *Logger) message(msg string) encoding {
	if len(msg) > 0 {
		b := buffer.Pool.Get()
		defer b.Free()
		b.Write(l.encoder.MessageBuffer.Bytes())
		l.encoder.BufferString(b, msg)
		return b.Bytes()
	}
	return nil
}
