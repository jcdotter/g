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

package env

const (
	Comment = '#'
	Sep     = '='
	Line    = '\n'
	Return  = '\r'
	Tab     = '\t'
	Esc     = '\\'
	Apost   = '\''
	Quote   = '"'
	Space   = ' '
	Export  = "export "
)

func search(b byte, in []byte, at int) (i int) {
	i = at
	for i < len(in) && in[i] != b {
		i++
	}
	return
}

func skip(b byte) bool {
	return b == Return || b == Line || b == Tab || b == Space
}

func next(b []byte, at int) int {
	for at < len(b) {
		switch c := b[at]; {
		case c == Export[0]:
			at = skipExport(b, at)
			fallthrough
		case c == Comment:
			at = skipComment(b, at)
		case skip(c):
			at++
		default:
			return at
		}
	}
	return at
}

func skipComment(b []byte, at int) int {
	if b[at] == Comment {
		return search(Line, b, at+1)
	}
	return at
}

func skipExport(b []byte, at int) int {
	if r := len(b) - at; r >= len(Export) && string(b[at:r]) == Export {
		at += len(Export)
	}
	return at
}

func parseKey(b []byte, at int) (key string, i int) {
	i = search(Sep, b, at)
	key = string(b[at:i])
	return
}

func parseValue(b []byte, at int) (val string, i int) {
	if c := b[at]; c == Apost || c == Quote {
		i = search(c, b, at+1)
		if b[i-1] != Esc {
			val = string(b[at+1 : i])
			i++
			return
		}
	}
	i = search(Line, b, at)
	val = string(b[at:i])
	return
}

func parse(b []byte) (vars map[string]string) {
	vars = make(map[string]string)
	var key, val string
	for at := next(b, 0); at < len(b); at = next(b, at) {
		if at < len(b) {
			key, at = parseKey(b, at)
			at = next(b, at+1)
			val, at = parseValue(b, at)
			vars[key] = val
		}
	}
	return
}

func parseItem(b []byte) (key, val string, i int) {
	key, i = parseKey(b, 0)
	val, i = parseValue(b, i+1)
	return
}
