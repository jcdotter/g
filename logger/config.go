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
	"io"
	"os"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/time"
)

var (
	defaultWriter io.Writer = os.Stdout
	defaultConfig *Config   = &Config{
		DefaultLevel: LevelInfo,
		LogTime:      true,
		LogPackage:   true,
		LogCaller:    true,
		LogFunction:  true,
	}
)

type Config struct {
	implemented  bool
	DefaultLevel Level
	LogTime      bool
	LogService   bool
	LogPackage   bool
	LogCaller    bool
	LogFunction  bool
	LogStatics   bool
	Fields       []*field
}

func NewConfig() *Config {
	return &Config{
		DefaultLevel: defaultConfig.DefaultLevel,
		LogTime:      defaultConfig.LogTime,
		LogPackage:   defaultConfig.LogPackage,
		LogCaller:    defaultConfig.LogCaller,
		LogFunction:  defaultConfig.LogFunction,
	}
}

// Build implements the logger configurations
func (l *Logger) Build() *Logger {
	l.Lock()
	defer l.Unlock()

	// build config
	if l.config == nil {
		l.config = NewConfig()
	} else if l.config.implemented {
		return l
	}
	l.config.implemented = true

	// build encoder
	if l.encoder == nil {
		l.encoder = NewEncoder()
	}
	l.encoder.PresetBuffers(l.config)

	// set writer
	if l.writers == nil {
		l.writers = append(make([]io.Writer, 0, 1), defaultWriter)
	}

	// build clock
	if l.config.LogTime {
		l.clock = time.GetClock().Format(l.encoder.TimeFmt)
		l.encoder.BufferBytes(l.encoder.TimeBuffer, l.clock.cache)
	}

	return l
}

// DefaultLevel sets the default log level
func (l *Logger) DefaultLevel(ll Level) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.DefaultLevel = ll
	return l
}

// Writers sets the writers
func (l *Logger) Writers(w ...io.Writer) *Logger {
	l.Lock()
	defer l.Unlock()
	l.writers = w
	return l
}

// AddWriter adds a writer
func (l *Logger) AddWriters(w ...io.Writer) *Logger {
	l.Lock()
	defer l.Unlock()
	if l.writers == nil {
		l.writers = append([]io.Writer{defaultWriter}, w...)
	} else {
		l.writers = append(l.writers, w...)
	}
	return l
}

// LevelKey sets the level key
func (l *Logger) LevelKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.LevelKey = s
	l.config.implemented = false
	return l
}

// LogTime sets whether to log time
func (l *Logger) LogTime(b bool) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogTime = b
	return l
}

// TimeFmt sets the time format
func (l *Logger) TimeFmt(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.TimeFmt = s
	if len(s) > 0 {
		l.config.LogTime = true
	}
	l.config.implemented = false
	return l
}

// TimeKey sets the time key
func (l *Logger) TimeKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.TimeKey = s
	l.config.implemented = false
	return l
}

// ServiceKey sets the service key
func (l *Logger) ServiceKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.ServiceKey = s
	if len(s) > 0 {
		l.config.LogService = true
	}
	l.config.implemented = false
	return l
}

// ServiceName sets the service name
func (l *Logger) ServiceName(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.ServiceName = s
	if len(s) > 0 {
		l.config.LogService = true
	}
	l.config.implemented = false
	return l
}

// CallIdKey sets the call id key
func (l *Logger) CallIdKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.CallIdKey = s
	l.config.implemented = false
	return l
}

// LogPackage sets whether to log package
func (l *Logger) LogPackage(b bool) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogPackage = b
	return l
}

// PackageKey sets the package key
func (l *Logger) PackageKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.PackageKey = s
	l.config.implemented = false
	return l
}

// LogCaller sets whether to log caller
func (l *Logger) LogCaller(b bool) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogCaller = b
	return l
}

// CallerKey sets the caller key
func (l *Logger) CallerKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.CallerKey = s
	l.config.implemented = false
	return l
}

// LogFunction sets whether to log function
func (l *Logger) LogFunction(b bool) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogFunction = b
	return l
}

// FunctionKey sets the function key
func (l *Logger) FunctionKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.FunctionKey = s
	l.config.implemented = false
	return l
}

// AddStaticField adds a static field to the logger
func (l *Logger) AddStaticField(name string, value any) *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogStatics = true
	if l.encoder.StaticBuffer == nil {
		l.encoder.StaticBuffer = buffer.Make(256)
	}
	l.encoder.BufferKey(l.encoder.StaticBuffer, l.encoder.ElemSep, name)
	l.encoder.BufferVal(l.encoder.StaticBuffer, value)
	return l
}

// RemoveStaticFields removes a static fields from the logger
func (l *Logger) RemoveStaticFields() *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.LogStatics = false
	l.encoder.StaticBuffer = nil
	return l
}

// AddField adds a field to the logger
func (l *Logger) AddField(name string, fn func(*Logger) any) *Logger {
	l.Lock()
	defer l.Unlock()
	if l.config.Fields == nil {
		l.config.Fields = make([]*field, 0, 4)
	}
	l.config.Fields = append(l.config.Fields, &field{
		name: name,
		pre:  []byte(name),
		fn:   fn,
	})
	return l
}

// RemoveField removes a field from the logger
func (l *Logger) RemoveField(name string) *Logger {
	l.Lock()
	defer l.Unlock()
	for i, f := range l.config.Fields {
		if f.name == name {
			l.config.Fields = append(l.config.Fields[:i], l.config.Fields[i+1:]...)
			return l
		}
	}
	return l
}

// RemoveFields removes all fields from the logger
func (l *Logger) RemoveFields() *Logger {
	l.Lock()
	defer l.Unlock()
	l.config.Fields = nil
	return l
}

// MessageKey sets the message key
func (l *Logger) MessageKey(s string) *Logger {
	l.Lock()
	defer l.Unlock()
	l.encoder.MessageKey = s
	l.config.implemented = false
	return l
}
