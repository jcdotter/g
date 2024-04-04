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
	"fmt"
	"io"
	"sync"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/time"
)

//	TODO:
//	- [ ] Service Writers
//	  - [ ] http
//	  - [ ] grpc...
//	- [ ] in app, pass logger and callid in context through middleware

// Logger is the logger struct containing the
// configurations and methods for logging
type Logger struct {
	sync.Mutex
	config  *Config     // logger configuration
	writers []io.Writer // multiwriter to Logger ouput(s)
	clock   *time.Clock // time clock
	encoder *Encoder    // encoder for log components
}

// New returns a new logger with provided options, if any
func New() *Logger {
	return new(Logger).Build()
}

// Write writes a log message to the logger
func (l *Logger) write(level Level, msg string, callid string, keyvals ...any) {
	if l.config == nil || !l.config.implemented {
		panic("logger not implemented")
	}
	b := buffer.Pool.Get()
	defer b.Free()
	b.WriteBytes(l.level(level))
	b.WriteBytes(l.time())
	b.WriteBytes(l.service())
	b.WriteBytes(l.callid(callid))
	b.WriteBytes(l.encCaller(2))
	b.WriteBytes(l.staticFields())
	b.WriteBytes(l.fields())
	b.WriteBytes(l.keyVals(keyvals...))
	b.WriteBytes(l.message(msg))
	b.WriteBytes(l.encoder.EndBuffer)

	// write buffer to writers concurrently
	switch len(l.writers) {
	case 0:
	case 1:
		l.writers[0].Write(b.Buffer())
	default:
		var wg sync.WaitGroup
		for _, w := range l.writers {
			wg.Add(1)
			go func(w io.Writer) {
				w.Write(b.Buffer())
				wg.Done()
			}(w)
		}
		wg.Wait()
	}
}

// format formats a log message
func (l *Logger) format(template string, args ...any) string {
	return fmt.Sprintf(template, args...)
}

// Write writes a log message to the logger
func (l *Logger) Write(level Level, msg string) {
	l.write(level, msg, "")
}

// Writef writes a template log message filled
// with args to the logger with the specified level
func (l *Logger) Writef(level Level, template string, args ...any) {
	l.write(level, l.format(template, args...), "")
}

// Writew writes a log message to the logger with
// with additional keyvalue pairs provided in args
func (l *Logger) Writew(level Level, msg string, keyvals ...any) {
	l.write(level, msg, "", keyvals...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.write(LevelDebug, msg, "")
}

// Debugf logs a debug message
func (l *Logger) Debugf(template string, args ...any) {
	l.write(LevelDebug, l.format(template, args...), "")
}

// Debugw logs a debug message
func (l *Logger) Debugw(msg string, keyvals ...any) {
	l.write(LevelDebug, msg, "", keyvals...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.write(LevelInfo, msg, "")
}

// Infof logs an info message
func (l *Logger) Infof(template string, args ...any) {
	l.write(LevelInfo, l.format(template, args...), "")
}

// Infow logs an info message
func (l *Logger) Infow(msg string, keyvals ...any) {
	l.write(LevelInfo, msg, "", keyvals...)
}

// Warn logs a warn message
func (l *Logger) Warn(msg string) {
	l.write(LevelWarn, msg, "")
}

// Warnf logs a warn message
func (l *Logger) Warnf(template string, args ...any) {
	l.write(LevelWarn, l.format(template, args...), "")
}

// Warnw logs a warn message
func (l *Logger) Warnw(msg string, keyvals ...any) {
	l.write(LevelWarn, msg, "", keyvals...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.write(LevelError, msg, "")
}

// Errorf logs an error message
func (l *Logger) Errorf(template string, args ...any) {
	l.write(LevelError, l.format(template, args...), "")
}

// Errorw logs an error message
func (l *Logger) Errorw(msg string, keyvals ...any) {
	l.write(LevelError, msg, "", keyvals...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(msg string) {
	l.write(LevelFatal, msg, "")
}

// Fatalf logs a fatal message
func (l *Logger) Fatalf(template string, args ...any) {
	l.write(LevelFatal, l.format(template, args...), "")
}

// Fatalw logs a fatal message
func (l *Logger) Fatalw(msg string, keyvals ...any) {
	l.write(LevelFatal, msg, "", keyvals...)
}

// Panic logs a panic message
func (l *Logger) Panic(msg string) {
	l.write(LevelPanic, msg, "")
}

// Panicf logs a panic message
func (l *Logger) Panicf(template string, args ...any) {
	l.write(LevelPanic, l.format(template, args...), "")
}

// Panicw logs a panic message
func (l *Logger) Panicw(msg string, keyvals ...any) {
	l.write(LevelPanic, msg, "", keyvals...)
}
