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

import (
	"strconv"
)

// Parser is a library of utilities for parsing []byte

// TODO:
// robust testing
// add parsers for:
//   - object: series of (key, value) pairs
//   - list: series of values
//   - key: series of characters
//   - value: series of characters

// ----------------------------------------------------------------------------
// CONDITIONAL STATEMENTS

// Condition stores a single conditional statement to be used
// for evaluating the next action in parsing a []byte. Conditions
// are used as the building blocks for the parser.
type Condition func(in []byte, at int) (ok bool, end int)
type Conditions []Condition

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

func cond(c bool, at, l int) (bool, int) {
	return c, at + l*BoolToInt(c)
}

// Cond returns a Condition with a value and an operator
// if no operator is provided, the Condition operator is =.
// Operators: (=, !, <, >, <=, >=).
// Example: Cond('x', '!') (not equal to "x")
func Cond(val byte, op ...byte) (c Condition) {
	switch condOp(op) {
	case 0:
		return func(in []byte, at int) (ok bool, end int) { return cond(val == in[at], at, 1) }
	case 1:
		return func(in []byte, at int) (ok bool, end int) { return cond(val != in[at], at, 1) }
	case 2:
		return func(in []byte, at int) (ok bool, end int) { return cond(val < in[at], at, 1) }
	case 3:
		return func(in []byte, at int) (ok bool, end int) { return cond(val > in[at], at, 1) }
	case 4:
		return func(in []byte, at int) (ok bool, end int) { return cond(val <= in[at], at, 1) }
	case 5:
		return func(in []byte, at int) (ok bool, end int) { return cond(val >= in[at], at, 1) }
	}
	return
}

// Cond returns a Condition with a value and an operator
// if no operator is provided, the Condition operator is =.
// Operators: (=, !, <, >, <=, >=).
// Example: CondString("hello", '!') (not equal to "hello")
func CondString(val string, op ...byte) (c Condition) {
	switch len(val) {
	case 0:
		panic("invalid Condition")
	case 1:
		return Cond(val[0], op...)
	}
	l := len(val)
	switch condOp(op) {
	case 0:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val == string(in[at:at+l]), at, l)
		}
	case 1:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val != string(in[at:at+l]), at, l)
		}
	case 2:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val < string(in[at:at+l]), at, l)
		}
	case 3:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val > string(in[at:at+l]), at, l)
		}
	case 4:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val <= string(in[at:at+l]), at, l)
		}
	case 5:
		return func(in []byte, at int) (ok bool, end int) {
			return cond(canExistString(val, in, at) && val >= string(in[at:at+l]), at, l)
		}
	}
	return
}

// OR comibines a series of Conditions into a single
// Condition validating if any of the Conditions are met.
// A Condition is a func(in []byte, at int) (ok bool, end int)
// that returns true if the Condition is met at the given int
// position in the provided []byte.
func OR(c ...Condition) Condition {
	return func(in []byte, at int) (ok bool, end int) {
		for _, i := range c {
			if ok, end = i(in, at); ok {
				return
			}
		}
		return
	}
}

// OR comibines a series of Conditions into a single
// Condition validating if all of the Conditions are met.
// A Condition is a func(in []byte, at int) (ok bool, end int)
// that returns true if the Condition is met at the given int
// position in the provided []byte.
func AND(c ...Condition) Condition {
	return func(in []byte, at int) (ok bool, end int) {
		for _, i := range c {
			if ok, end = i(in, at); !ok {
				return
			}
		}
		return
	}
}

// NOT returns a Condition that returns true if the provided
// Condition is not met at the given int position in the provided
// []byte.
func NOT(c Condition) Condition {
	return func(in []byte, at int) (ok bool, end int) {
		if ok, end = c(in, at); !ok {
			return true, end
		}
		return false, at
	}
}

// And is used to evaluate a series of Conditions with a logical AND operator
func (c Conditions) And(in []byte, at int) (ok bool, end int) {
	return AND(c...)(in, at)
}

// Or is used to evaluate a series of conitions with a logical OR operator
func (c Conditions) Or(in []byte, at int) (ok bool, end int) {
	return OR(c...)(in, at)
}

// Not is used to evaluate a single Condition with a logical NOT operator
func (c Conditions) Not(in []byte, at int) (ok bool, end int) {
	return NOT(c[0])(in, at)
}

// ----------------------------------------------------------------------------
// PARSERS

// Parser is a function that parses an item in a []byte at the position
// provided and returns the item and the position after the item.
type Parser func(in []byte, at int) (item []byte, end int)

// item represents a parsable item where parsing begins when a pre-check
// is met and ends when a post-check is met.
type item struct {
	pre      Condition                              // syntax pre check Conditions: how to find the item
	post     Condition                              // syntax post check Conditions: how to end the item
	encl     []item                                 // enclosed items: items to be escaped/skipped if found
	precall  func(i *item, in []byte, at int) error // pre-call function called on pre-check success
	postcall func(i *item, in []byte, at int) error // post-call function called on post-check success
}

// Item returns a new parsable item with pre and post checks
// and a list of enclosed items to be escaped/skipped if found
func Item(pre, post Condition, encl ...item) (i item) {
	return item{pre: pre, post: post, encl: encl}
}

// Exists returns true if the item prefix condition
// is true in the provided []byte  at the specified
// position, otherwise it returns false.
func (i *item) Exists(in []byte, at int) (ok bool, end int) {
	return i.pre(in, at)
}

// Search returns the next position of the item prefix
// in the provided []byte at or after the specified
// position. If the item prefix does not exist, the
// returns -1.
func (i *item) Search(in []byte, at int) int {
	for at < len(in) {
		if ok, _ := i.pre(in, at); ok {
			return at
		}
		for _, e := range i.encl {
			if _, o := e.Parse(in, at); o > at {
				at = o
				goto next
			}
		}
		at++
	next:
	}
	return -1
}

// Parse returns the item in the provided []byte at the
// specified position. If the item exists, the function
// returns the item and the position after the item, otherwise
// it returns nil and the original position.
func (i *item) Parse(in []byte, at int) (out []byte, end int) {
	var ok bool
	if ok, end = i.pre(in, at); ok {
		i.execPrecall(in, at)
		at = end
		for end < len(in) {
			if ok, n := i.post(in, end); ok {
				out = in[at:end]
				i.execPostcall(in, n)
				end = n
				return
			}
			for _, j := range i.encl {
				if _, n := j.Parse(in, end); n > end {
					end = n
					goto next
				}
			}
			end++
		next:
		}
	}
	return nil, at
}

func (i *item) execPrecall(in []byte, at int) error {
	if i.precall != nil {
		return i.precall(i, in, at)
	}
	return nil
}

func (i *item) execPostcall(in []byte, at int) error {
	if i.postcall != nil {
		return i.postcall(i, in, at)
	}
	return nil
}

// ----------------------------------------------------------------------------
// ITEM LIBRARY

var (
	CommentItem      = Item(CondString("//"), Cond('\n'))
	BlockCommentItem = Item(CondString("/*"), CondString("*/"))
	StringItem       = item{
		pre: IsQuote,
		precall: func(i *item, in []byte, at int) error {
			q := in[at]
			i.post = func(in []byte, at int) (ok bool, end int) {
				return cond(in[at] == q && in[at-1] != '\\', at, 1)
			}
			return nil
		},
	}
	ModuleItem = Item(
		CondString("module "),                               // precheck
		OR(NOT(IsChar), CondString("//"), CondString("/*")), // postcheck
		CommentItem, BlockCommentItem, StringItem, // enclosed items
	)
	VersionItem = Item(
		CondString("version "),                              // precheck
		OR(NOT(IsChar), CondString("//"), CondString("/*")), // postcheck
		CommentItem, BlockCommentItem, StringItem, // enclosed items
	)
)

// ----------------------------------------------------------------------------
// PARSE LIBRARY

func isChar(b byte) bool                    { return b > 0x20 && b < 0x7e }
func isNum(b byte) bool                     { return b > 0x29 && b < 0x3a }
func isAlpha(b byte) bool                   { return (b > 0x40 && b < 0x5b) || (b > 0x60 && b < 0x7b) }
func isAlphamNum(b byte) bool               { return isAlpha(b) || isNum(b) }
func isQuote(b byte) bool                   { return b == 0x22 || b == 0x27 || b == 0x60 }
func canExist(item, in []byte, at int) bool { return len(item) != 0 && len(item) <= len(in)-at }
func canExistString(item string, in []byte, at int) bool {
	return len(item) != 0 && len(item) <= len(in)-at
}

// IsChar checks if in byte at the specified position in the provided
// []byte is a non-whitspace ('\r','\n','\t',' ') printable character
func IsChar(in []byte, at int) (ok bool, end int) {
	return isChar(in[at]), at + 1
}

// IsNum checks if in byte at the specified position in the provided
// []byte is a number (0-9)
func IsNum(in []byte, at int) (ok bool, end int) {
	return isNum(in[at]), at + 1
}

// IsAlpha checks if in byte at the specified position in the provided
// []byte is a letter (a-z, A-Z)
func IsAlpha(in []byte, at int) (ok bool, end int) {
	return isAlpha(in[at]), at + 1
}

// IsAlphamNum checks if in byte at the specified position in the provided
// []byte is a letter (a-z, A-Z) or a number (0-9)
func IsAlphamNum(in []byte, at int) (ok bool, end int) {
	return isAlphamNum(in[at]), at + 1
}

// IsQuote checks if in byte at the specified position in the provided
// []byte is a quote (",',`)
func IsQuote(in []byte, at int) (ok bool, end int) {
	return isQuote(in[at]), at + 1
}

// Exists checks if the provided []byte item exists in the provided []byte at the
// specified position. If the item exists, the function returns true, otherwise
// it returns false.
func Exists(item, in []byte, at int) (ok bool, end int) {
	return canExist(item, in, at) &&
		string(item) == string(in[at:at+len(item)]), at
}

// ExistsString checks if the provided string exists in the provided []byte at the
// specified position. If the item exists, the function returns true, otherwise
// it returns false.
func ExistsString(item string, in []byte, at int) (ok bool, end int) {
	return len(item) != 0 && len(item) <= len(in)-at &&
		item == string(in[at:at+len(item)]), at
}

// Search checks if the provided []byte item exists in the provided []byte at
// or after the specified position. If the item exists, the function returns
// the position of the item, otherwise it returns -1.
func Search(item, in []byte, at int) (found bool, foundAt int) {
	if len(item) == 1 {
		return Find(item[0], in, at)
	}
	if canExist(item, in, at) {
		for foundAt = at; foundAt < len(in); foundAt++ {
			if found, foundAt = Find(item[0], in, foundAt); found {
				if is, _ := Exists(item, in, foundAt); is {
					return
				}
			}
		}
	}
	foundAt = at
	return
}

// Find checks if the provided byte exists in the provided []byte at or after
// the specified position. If the byte exists, the function returns the position
// of the byte, otherwise it returns -1.
func Find(b byte, in []byte, at int) (found bool, foundAt int) {
	for foundAt = at; foundAt < len(in); foundAt++ {
		if found = in[foundAt] == b; found {
			return
		}
	}
	foundAt = at
	return
}

func Next(c Condition, in []byte, at int) (found bool, foundAt int) {
	for foundAt = at; foundAt < len(in); foundAt++ {
		if found, foundAt = c(in, foundAt); found {
			return
		}
	}
	foundAt = at
	return
}

// StringLit checks if a literal string exists at the specified position in the
// provided []byte. If the provided []byte contains a literal string, the function
// returns the literal string and the position after the literal string, otherwise
// it returns nil and -1.
func StringLit(in []byte, at int) (s []byte, end int) {
	if q := in[at]; isQuote(q) {
		for i := at + 1; i < len(in); i++ {
			if in[i] == q && in[i-1] != '\\' {
				i++
				return in[at:i], i
			}
		}
	}
	return nil, -1
}

// Num checks if a number exists at the specified position in the provided
// []byte. If the provided []byte contains a number, the function returns the
// number and the position after the number, otherwise it returns nil and -1.
func Num(in []byte, at int) (n []byte, end int) {
	c := in[at]
	d, e, i := 0, 0, at
	if c == '-' || c == '+' {
		i++
		c = in[i]
	}
	if isNum(c) {
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
	return nil, at
}

// Bool checks if a boolean exists at the specified position in the provided
// []byte. If the provided []byte contains a boolean, the function returns the
// boolean and the position after the boolean, otherwise it returns nil and -1.
func Bool(in []byte, at int) (n []byte, end int) {
	if exists, _ := ExistsString("true", in, at); exists {
		return []byte("true"), at + 4
	}
	if exists, _ := ExistsString("false", in, at); exists {
		return []byte("false"), at + 5
	}
	return nil, at
}

// Null checks if a null exists at the specified position in the provided
// []byte. If the provided []byte contains a null, the function returns the
// null and the position after the null, otherwise it returns nil and -1.
func Null(in []byte, at int) (n []byte, end int) {
	if exists, _ := ExistsString("null", in, at); exists {
		return []byte("null"), at + 4
	}
	if exists, _ := ExistsString("nil", in, at); exists {
		return []byte("nil"), at + 3
	}
	return nil, at
}

func Module(in []byte, at int) (n []byte, end int) {
	return ModuleItem.Parse(in, at)
}

func Version(in []byte, at int) (n []byte, end int) {
	return VersionItem.Parse(in, at)
}

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
	if f := Float(s); f != 0 {
		return int(f)
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

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
