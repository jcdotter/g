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
	LevelKeyBuffer []byte // eg. `{"level":` (every log entry must have a level)
	LevelBuffer    []byte // eg. `{"level":"info"`
	TimeKeyBuffer  []byte // eg. `,"ts":`
	TimeBuffer     []byte // eg. `,"ts":"2006-01-02 15:04:05.000"`
	ServiceBuffer  []byte // eg. `,"svc":"my-service"` (service name doesn't change)
	CallIdBuffer   []byte // eg. `,"callid":`
	PackageBuffer  []byte // eg. `,"pkg":`
	CallerBuffer   []byte // eg. `,"src":`
	FunctionBuffer []byte // eg. `,"fn":`
	StaticBuffer   []byte // eg. `,"key":"value"`
	MessageBuffer  []byte // eg. `,"msg":`
	EndBuffer      []byte // eg. `}`
	LogStart       byte   // eg. `{`
	LogEnd         byte   // eg. `}`
	LogSep         byte   // eg. `\n`
	ElemSep        byte   // eg. `,`
	KeyValSep      byte   // eg. `:`
	Quote          byte   // eg. `"`
	TimeFmt        string
	CallerCache    map[uintptr][]byte
}

// CreateKey creates a key for the encoder
func (e *Encoder) BufferKey(key string) []byte {
	return append(append([]byte{e.ElemSep, e.Quote}, []byte(key)...), e.Quote, e.KeyValSep)
}

func (e *Encoder) BufferVal(val any) []byte {
	j, _ := json.Marshal(val)
	return j
}

// BufferString creates a string for the encoder
func (e *Encoder) BufferString(s string) []byte {
	return append(append([]byte{e.Quote}, []byte(s)...), e.Quote)
}

// BufferBytes creates a byte slice for the encoder
func (e *Encoder) BufferBytes(b []byte) []byte {
	return append(append([]byte{e.Quote}, b...), e.Quote)
}

// CreateKeyVal creates a key value pair for the encoder
func (e *Encoder) BufferKeyVal(key string, val any) []byte {
	return append(e.BufferKey(key), e.BufferVal(val)...)
}

// BufferKeyValString creates a key value pair for the encoder
func (e *Encoder) BufferKeyValString(key string, val string) []byte {
	return append(e.BufferKey(key), e.BufferString(val)...)
}

func (e *Encoder) BufferLevels() {
	b := buffer.Pool.Get()
	defer b.Free()
	e.LevelIndex = make([]uint8, len(levelNameIndex))
	for i := 0; i < len(levelNameIndex)-1; i++ {
		b.WriteBytes(e.LevelKeyBuffer)
		b.WriteByte(e.Quote)
		b.WriteString(Level(i).String())
		b.WriteByte(e.Quote)
		e.LevelIndex[i+1] = uint8(b.Len())
	}
	e.LevelBuffer = b.Bytes()
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
	return l.encoder.LevelBuffer[l.encoder.LevelIndex[ll]:l.encoder.LevelIndex[ll+1]]
}

func (l *Logger) time() encoding {
	if l.config.LogTime {
		if l.clock.refresh() {
			l.encoder.TimeBuffer = append(l.encoder.TimeKeyBuffer, l.encoder.BufferBytes(l.clock.cache)...)
		}
		return l.encoder.TimeBuffer
	}
	return nil
}

func (l *Logger) service() encoding {
	if l.config.LogService {
		return l.encoder.ServiceBuffer
	}
	return nil
}

func (l *Logger) callid(cid string) encoding {
	if len(cid) > 0 {
		return append(l.encoder.CallIdBuffer, l.encoder.BufferString(cid)...)
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
			if l.config.LogPackage {
				b.WriteBytes(l.encoder.PackageBuffer)
				b.WriteBytes(l.encoder.BufferString(caller.Pkg().Path))
			}
			if l.config.LogCaller {
				b.WriteBytes(l.encoder.CallerBuffer)
				cb := buffer.Pool.Get()
				defer cb.Free()
				cb.WriteString(caller.File().Name)
				cb.WriteByte(':')
				cb.WriteInt(caller.Line())
				b.WriteBytes(l.encoder.BufferBytes(cb.Bytes()))
			}
			if l.config.LogFunction {
				b.WriteBytes(l.encoder.FunctionBuffer)
				b.WriteBytes(l.encoder.BufferString(caller.Func().Name))
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
		return l.encoder.StaticBuffer
	}
	return nil
}

func (l *Logger) fields() encoding {
	if len(l.config.Fields) > 0 {
		b := buffer.Pool.Get()
		defer b.Free()
		for _, f := range l.config.Fields {
			b.WriteBytes(f.pre)
			b.WriteBytes(l.encoder.BufferVal(f.fn(l)))
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
			b.WriteBytes(l.encoder.BufferKeyVal(keyvals[i].(string), keyvals[i+1]))
		}
		return b.Bytes()
	}
	return nil
}

func (l *Logger) message(msg string) encoding {
	if len(msg) > 0 {
		return append(l.encoder.MessageBuffer, l.encoder.BufferString(msg)...)
	}
	return nil
}
