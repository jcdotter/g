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

package inspect

import "errors"

// -----------------------------------------------------------------------------
// ERRORS

// Error types.
var (
	ErrEOF                 = errors.New("end of file")
	ErrInvalidKeyword      = errors.New("invalid keyword")
	ErrInvalidCharacter    = errors.New("invalid character")
	ErrInvalidOperator     = errors.New("invalid operator")
	ErrInvalidEntity       = errors.New("invalid entity: must be const, var, type, or func")
	ErrIncompleteEnclosure = errors.New("incomplete enclosure")
	ErrNotType             = errors.New("no type specified")
	ErrInvalidType         = errors.New("invalid type")
	ErrNotExist            = errors.New("does not exist")
)

// -----------------------------------------------------------------------------
// KEYWORDS

// go declaration keywords
var (
	CommentChar byte = '/'
	ConstChar   byte = 'c'
	FuncChar    byte = 'f'
	ImportChar  byte = 'i'
	PackageChar byte = 'p'
	TypeChar    byte = 't'
	VarChar     byte = 'v'

	CommentLineKey  = []byte("//")
	CommentBlockKey = []byte("/*")
	CommentCloseKey = []byte("*/")
	ConstKey        = []byte("const ")
	FuncKey         = []byte("func ")
	ImportKey       = []byte("import ")
	PackageKey      = []byte("package ")
	TypeKey         = []byte("type ")
	VarKey          = []byte("var ")
)

// go enclosure characters

var (
	OpenParen    byte = '(' // open parenthesis or round bracket '('
	CloseParen   byte = ')' // close parenthesis or round bracket ')'
	OpenBrace    byte = '{' // open brace or curly bracket '{'
	CloseBrace   byte = '}' // close brace or curly bracket '}'
	OpenBracket  byte = '[' // open bracket or square bracket '['
	CloseBracket byte = ']' // close bracket or square bracket ']'

	Paren   = Encl{OpenParen, CloseParen}     // Paren is a pair of round brackets
	Brace   = Encl{OpenBrace, CloseBrace}     // Brace is a pair of curly brackets
	Bracket = Encl{OpenBracket, CloseBracket} // Bracket is a pair of square brackets
)

type Encl struct{ Open, Close byte }

// go type keywords

var (
	ElipsKey     = []byte("...")
	SliceKey     = []byte("[]")
	MapKey       = []byte("map")
	FnKey        = []byte("func")
	StructKey    = []byte("struct")
	InterfaceKey = []byte("interface")
	ChanKey      = []byte("chan ")
	ChanWOKey    = []byte("chan<- ")
	ChanROKey    = []byte("<-chan ")
)

var (
	TB byte = '\t' // tab '\t'
	LF byte = '\n' // line feed '\n'
	CR byte = '\r' // carriage return '\r'
	SP byte = ' '  // space ' '
)

// go delimiters
var (
	Comma byte = ','
	Dot   byte = '.'
	Colon byte = ':'
	Semi  byte = ';'
)

// go operators
var (
	Assign byte = '='
	Not    byte = '!'
	Plus   byte = '+'
	Minus  byte = '-'
	Star   byte = '*'
	Slash  byte = '/'
	Mod    byte = '%'
	And    byte = '&'
	Or     byte = '|'
	Xor    byte = '^'
	ShiftL byte = '<'
	ShiftR byte = '>'
)
