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

package parser

import "strconv"

type Parser struct {
	b []byte // buffer
	c int    // cursor
}

// object: series of (key, value) pairs
// list: series of values
// key: series of characters
// value: series of characters
// keywords

// Statndard Items
// comment, whitestpace, string, number, boolean, null

// if Find(Comment.Pre()).In(b).At(c) {

// Exists(item []byte, in []byte, at int) (int, bool)
// Search(item []byte, in []byte, at int) int

type condition struct {
	o byte   // operator (=, !, <, >, <=, >=)
	b []byte // value
}

// Cond returns a condition with a value and an operator
// if no operator is provided, the condition operator is =.
// Operators: =, !, <, >, <=, >=
func Cond(val []byte, op ...byte) (c condition) {
	c = condition{b: val, o: condOp(op)}
	if len(c.b) == 0 || (c.o > 1 && len(c.b) != 1) {
		panic("invalid condition")
	}
	return
}

func condOp(op []byte) byte {
	switch string(op) {
	case "", "=", "==":
		return 0
	case "!", "!=":
		return 1
	case "<":
		return 2
	case ">":
		return 3
	case "<=":
		return 4
	case ">=":
		return 5
	}
	panic("invalid operator")
}

type checks []check
type check func(in []byte, at int) bool

func Checks(item ...condition) (c checks) {
	for _, i := range item {
		if len(i.b) == 1 {
			switch i.o {
			case 0:
				c = append(c, func(in []byte, at int) bool { return in[at] == i.b[0] })
			case 1:
				c = append(c, func(in []byte, at int) bool { return in[at] != i.b[0] })
			case 2:
				c = append(c, func(in []byte, at int) bool { return in[at] < i.b[0] })
			case 3:
				c = append(c, func(in []byte, at int) bool { return in[at] > i.b[0] })
			case 4:
				c = append(c, func(in []byte, at int) bool { return in[at] <= i.b[0] })
			case 5:
				c = append(c, func(in []byte, at int) bool { return in[at] >= i.b[0] })
			}
		} else {
			c = append(c, func(in []byte, at int) bool { return Exists(i.b, in, at) == (i.o == 0) })
		}
	}
	return
}

type Value struct {
	pre checks
	suf checks
}

// ----------------------------------------------------------------------------
// PARSE LIBRARY

func isChar(b byte) bool                    { return b > 0x20 && b < 0x7e }
func isNum(b byte) bool                     { return b > 0x29 && b < 0x3a }
func isAlpha(b byte) bool                   { return (b > 0x40 && b < 0x5b) || (b > 0x60 && b < 0x7b) }
func isAlphamNum(b byte) bool               { return isAlpha(b) || isNum(b) }
func isQuote(b byte) bool                   { return b == 0x22 || b == 0x27 || b == 0x60 }
func canExist(item, in []byte, at int) bool { return len(item) != 0 && len(item) <= len(in)-at }

// IsChar checks if in byte at the specified position in the provided
// []byte is a non-whitspace ('\r','\n','\t',' ') printable character
func IsChar(in []byte, at int) bool {
	return isChar(in[at])
}

// IsNum checks if in byte at the specified position in the provided
// []byte is a number (0-9)
func IsNum(in []byte, at int) bool {
	return isNum(in[at])
}

// IsAlpha checks if in byte at the specified position in the provided
// []byte is a letter (a-z, A-Z)
func IsAlpha(in []byte, at int) bool {
	return isAlpha(in[at])
}

// IsAlphamNum checks if in byte at the specified position in the provided
// []byte is a letter (a-z, A-Z) or a number (0-9)
func IsAlphamNum(in []byte, at int) bool {
	return isAlphamNum(in[at])
}

// IsQuote checks if in byte at the specified position in the provided
// []byte is a quote (",',`)
func IsQuote(in []byte, at int) bool {
	return isQuote(in[at])
}

// Exists checks if the provided []byte item exists in the provided []byte at the
// specified position. If the item exists, the function returns true, otherwise
// it returns false.
func Exists(item, in []byte, at int) bool {
	return canExist(item, in, at) &&
		string(item) == string(in[at:at+len(item)])
}

// Search checks if the provided []byte item exists in the provided []byte at
// or after the specified position. If the item exists, the function returns
// the position of the item, otherwise it returns -1.
func Search(item, in []byte, at int) int {
	if len(item) == 1 {
		return Find(item[0], in, at)
	}
	if canExist(item, in, at) {
		for i := at; i < len(in); i++ {
			if i = Find(item[0], in, at); Exists(item, in, i) {
				return i
			}
		}
	}
	return -1
}

// Find checks if the provided byte exists in the provided []byte at or after
// the specified position. If the byte exists, the function returns the position
// of the byte, otherwise it returns -1.
func Find(b byte, in []byte, at int) int {
	for i := at; i < len(in); i++ {
		if in[i] == b {
			return i
		}
	}
	return -1
}

// StringLit checks if a literal string exists at the specified position in the
// provided []byte. If the provided []byte contains a literal string, the function
// returns the literal string and the position after the literal string, otherwise
// it returns nil and -1.
func StringLit(in []byte, at int) (s []byte, end int) {
	if q := in[at]; isQuote(q) {
		for i := at + 1; i < len(in); i++ {
			if in[i] == q && in[i-1] != '\\' {
				//i++
				return in[at:i], i
			}
		}
	}
	return nil, -1
}

// Number checks if a number exists at the specified position in the provided
// []byte. If the provided []byte contains a number, the function returns the
// number and the position after the number, otherwise it returns nil and -1.
func Number(in []byte, at int) (n []byte, end int) {
	c := in[at]
	d, e := 0, 0
	if c == '-' || c == '+' {
		at++
		c = in[at]
	}
	if isNum(c) {
		i := at
		for ; i < len(in); i++ {
			c := in[i]
			// handle decimal
			if c == '.' {
				if d == 0 && isNum(in[i+1]) {
					d++
					i++
					continue
				}
				break
			}
			// handle exponent
			if c == 'e' || c == 'E' {
				if c = in[i+1]; c == '-' || c == '+' {
					i++
				}
				if isNum(in[i+1]) {
					e++
					i++
					continue
				}
				break
			}
			// handle non-numbers
			if !isNum(c) {
				break
			}
		}
		return in[at:i], i
	}
	return nil, -1
}

func Boolean() {}
func Null()    {}

// ----------------------------------------------------------------------------
// CONVERSIONS

// String converts the provided []byte of a literal string to a string
func String(s []byte) string {
	if len(s) == 0 {
		return ""
	}
	if q := s[0]; isQuote(q) && q == s[len(s)-1] {
		return string(s[1 : len(s)-1])
	}
	return string(s)
}

// Int converts the provided []byte of a number to an int
func Int(s []byte) int {
	if n, err := strconv.Atoi(string(s)); err == nil {
		return n
	}
	return 0
}

// Float converts the provided []byte of a number to a float64
func Float(s []byte) float64 {
	if n, err := strconv.ParseFloat(string(s), 64); err == nil {
		return n
	}
	return 0
}
