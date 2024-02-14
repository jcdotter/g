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

// Level is a log level
type Level uint8

const (
	// LevelDebug is the debug level
	LevelDebug Level = iota
	// LevelInfo is the info level
	LevelInfo
	// LevelWarn is the warn level
	LevelWarn
	// LevelError is the error level
	LevelError
	// LevelFatal is the fatal level
	LevelFatal
	// LevelPanic is the panic level
	LevelPanic
)

const levelName = `debuginfowarnerrorfatalpanic`

var levelNameIndex = [...]uint8{0, 5, 9, 13, 18, 23, 28}

// String returns the string representation of a log level
func (l Level) String() string {
	return levelName[levelNameIndex[l]:levelNameIndex[l+1]]
}
