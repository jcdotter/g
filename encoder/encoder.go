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

package encoder

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/jcdotter/go/buffer"
	t "github.com/jcdotter/go/typ"
	"github.com/jcdotter/go/uuid"
)

// ----------------------------------------------------------------------------
// Speed considerations:
// 1. Pass structure parts to children and append indents,
//    rather than using ancestry
// 2. encoding pkg does not handle recursive structures

// ----------------------------------------------------------------------------
// ENCODER IMPLEMENTATION
// a generic encoder for serializing and deserializing
// golang values to and from []byte in a specific format

type Encoder struct {
	// encoder state
	buffer      *buffer.Buffer // the encoder buffer
	value       any            // the value being encodeled
	cursor      int            // the current position in the buffer
	curDepth    int            // the current depth of the data structure
	curIndent   int            // the current indentation level
	hasBrackets bool           // true when encoding with brackets around data objects
	// encoding syntax
	Type              string // the type of encoder. json, yaml, etc.
	Space             []byte // the space characters
	LineBreak         []byte // the line break characters
	Indent            []byte // the indentation characters
	Quote             []byte // the quote characters, first is default, additional are alternate
	Escape            []byte // the escape characters for string special char escape
	Null              []byte // the null value
	ValEnd            []byte // the characters that separate values
	KeyEnd            []byte // the characters that separate keys and values
	BlockCommentStart []byte // the characters that start a block comment
	BlockCommentEnd   []byte // the characters that end a block comment
	LineCommentStart  []byte // the characters that start a single line comment
	LineCommentEnd    []byte // the characters that end a single line comment
	SliceStart        []byte // the characters that start a slice or array
	SliceEnd          []byte // the characters that end a slice or array
	SliceItem         []byte // the characters before each slice element
	MapStart          []byte // the characters that start a hash map
	MapEnd            []byte // the characters that end a hash map
	InlineSyntax      *InlineSyntax
	// encoding flags
	Format           bool // when true, encode with formatting, indentation, and line breaks
	FormatWithSpaces bool // when true, encode with space between keys and values
	CascadeOnlyDeep  bool // when true, encode single-depth slices and maps with inline syntax
	QuotedKey        bool // when true, encode map keys with quotes
	QuotedString     bool // when true, encode strings with quotes
	QuotedSpecial    bool // when true, encode strings with quotes if they contain special characters
	QuotedNum        bool // when true, encode numbers with quotes
	QuotedBool       bool // when true, encode bools with quotes
	QuotedNull       bool // when true, encode null with quotes
	RecursiveName    bool // when true, include name, string or type of recursive value in encoding, otherwise, exclude all recursion
	DecodeTyped      bool // when true, decode to typed values (int, float64, bool, string) instead of just strings
	EncodeMethods    bool // when true, encode structs with a Encode method by calling the method
	ExcludeZeros     bool // when true, exclude zero and nil values from encoding
	// encoder cache
	space      byte
	quote      byte
	escape     byte
	valEnd     []byte
	keyEnd     []byte
	ivalEnd    []byte
	sliceParts map[string][3][]byte
	mapParts   map[string][3][]byte
	hasTag     map[*t.Type]bool
	tagKeys    map[*t.Type][]string
	methods    map[*t.Type]int
}

type InlineSyntax struct {
	ValEnd     []byte // the characters that separate values
	SliceStart []byte // the characters that start a slice or array
	SliceEnd   []byte // the characters that end a slice or array
	MapStart   []byte // the characters that start a hash map
	MapEnd     []byte // the characters that end a hash map
}

type ancestor struct {
	typ     *t.Type
	pointer uintptr
}

var (
	sliceType = t.TypeOf([]any(nil))
	mapType   = t.TypeOf(map[string]any(nil))
)

// ----------------------------------------------------------------------------
// PRESET ENCODERS
// JSON, YAML...

var (
	Json = &Encoder{
		Type:              "json",
		FormatWithSpaces:  true,
		CascadeOnlyDeep:   true,
		QuotedKey:         true,
		QuotedString:      true,
		ExcludeZeros:      true,
		Space:             []byte(" \t\n\v\f\r"),
		Indent:            []byte("  "),
		Quote:             []byte(`"`),
		Escape:            []byte(`\`),
		ValEnd:            []byte(","),
		KeyEnd:            []byte(":"),
		BlockCommentStart: []byte("/*"),
		BlockCommentEnd:   []byte("*/"),
		LineCommentStart:  []byte("//"),
		LineCommentEnd:    []byte("\n"),
		SliceStart:        []byte("["),
		SliceEnd:          []byte("]"),
		MapStart:          []byte("{"),
		MapEnd:            []byte("}"),
	}
	Yaml = &Encoder{
		Type:             "yaml",
		Format:           true,
		FormatWithSpaces: true,
		QuotedSpecial:    true,
		ExcludeZeros:     true,
		Space:            []byte(" \t\v\f\r"),
		Indent:           []byte("  "),
		Quote:            []byte(`"'`),
		Escape:           []byte(`\`),
		KeyEnd:           []byte(":"),
		LineCommentStart: []byte("#"),
		LineCommentEnd:   []byte("\n"),
		SliceItem:        []byte("- "),
		InlineSyntax: &InlineSyntax{
			ValEnd:     []byte(", "),
			SliceStart: []byte("["),
			SliceEnd:   []byte("]"),
			MapStart:   []byte("{"),
			MapEnd:     []byte("}"),
		},
	}
)

// ----------------------------------------------------------------------------
// Init Utilities
// methods for setting up the encoder

func init() {
	Json.Init()
	Yaml.Init()
}

func (m *Encoder) Init() {
	m.Reset()
	if m.Null == nil {
		m.Null = []byte("null")
	}
	if m.Space == nil {
		m.Space = []byte(" \t\n\v\f\r")
	}
	if m.Quote == nil && (m.QuotedString || m.QuotedKey || m.QuotedSpecial || m.QuotedNum || m.QuotedBool || m.QuotedNull) {
		m.Quote = []byte(`"`)
	}
	if m.Escape == nil && m.Quote != nil {
		m.Escape = []byte(`\`)
	}
	m.hasBrackets = !(m.MapStart == nil || m.MapEnd == nil || m.SliceStart == nil || m.SliceEnd == nil)
	if !m.hasBrackets && !m.Format {
		panic("cannot encode without brackets or formatting: unable to determine data structure")
	}
	if m.SliceItem != nil && m.hasBrackets {
		panic("slice item reserved for bracketless encoding")
	}
	if !m.hasBrackets && InBytes('\n', m.Space) {
		panic("cannot encode without brackets when line breaks in space characters")
	}
	m.space = m.Space[0]
	m.quote = m.Quote[0]
	m.escape = m.Escape[0]
	m.keyEnd = m.KeyEnd
	m.valEnd = m.ValEnd
	m.sliceParts = map[string][3][]byte{}
	m.mapParts = map[string][3][]byte{}
	m.initFormat()
}

func (m *Encoder) initFormat() {
	if m.Format {
		if m.Indent == nil {
			m.Indent = []byte("  ")
		}
		if m.LineBreak == nil {
			m.LineBreak = []byte("\n")
		}
		if m.InlineSyntax == nil {
			m.InlineSyntax = &InlineSyntax{
				ValEnd:     m.ValEnd,
				SliceStart: m.SliceStart,
				SliceEnd:   m.SliceEnd,
				MapStart:   m.MapStart,
				MapEnd:     m.MapEnd,
			}
		}
		if m.FormatWithSpaces {
			m.keyEnd = append(m.KeyEnd, m.space)
			m.valEnd = append(m.ValEnd, m.space)
			m.ivalEnd = append(m.InlineSyntax.ValEnd, m.space)
		}
	}
}

func (m *Encoder) New() *Encoder {
	n := *m
	if m.InlineSyntax != nil {
		s := *m.InlineSyntax
		n.InlineSyntax = &s
	}
	return &n
}

func (m *Encoder) Reset() {
	if m.buffer == nil {
		m.buffer = buffer.New()
	} else {
		m.buffer.Reset()
	}
	m.cursor = 0
	m.curIndent = 0
	m.value = nil
	m.sliceParts = map[string][3][]byte{}
	m.mapParts = map[string][3][]byte{}
	m.hasTag = nil
	m.tagKeys = nil
	m.methods = nil
}

func (m *Encoder) ResetCursor() {
	m.cursor = 0
	m.curDepth = 0
	m.curIndent = 0
}

// ----------------------------------------------------------------------------
// Read Utilities
// for getting non-exported state values from the encoder

func (m *Encoder) Buffer() []byte {
	return m.buffer.Buffer()
}

func (m *Encoder) Cursor() int {
	return m.cursor
}

func (m *Encoder) CurIndent() int {
	return m.curIndent
}

func (m *Encoder) Len() int {
	return m.buffer.Len()
}

func (m *Encoder) Value() any {
	return m.value
}

func (m Encoder) Bytes() []byte {
	return m.buffer.Bytes()
}

func (m Encoder) String() string {
	return m.buffer.String()
}

func (m Encoder) Formatted(indent ...int) string {
	difIndent := false
	oldIndent := m.Indent
	oldBreak := m.LineBreak
	m.initFormat()
	if len(indent) > 0 {
		in := bytes.Repeat([]byte(" "), indent[0])
		if !bytes.Equal(m.Indent, in) {
			m.Indent = in
			difIndent = true
		}
	}
	if (difIndent || m.buffer == nil || !m.Format || !bytes.Equal(oldBreak, m.LineBreak)) && m.value != nil {
		oldFormat := m.Format
		m.Format = true
		m.initFormat()
		m.ResetCursor()
		m.Encode(m.value)
		m.Indent = oldIndent
		m.Format = oldFormat
	}
	m.LineBreak = oldBreak
	return m.String()
}

func (m Encoder) Map() map[string]any {
	v := t.ValueOf(m.value)
	switch v.Kind() {
	case t.MAP:
		return m.value.(map[string]any)
	case t.SLICE:
		return v.Slice().Map()
	}
	return nil
}

func (m Encoder) Slice() []any {
	v := t.ValueOf(m.value)
	switch v.Kind() {
	case t.MAP:
		return v.Map().Values()
	case t.SLICE:
		return m.value.([]any)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Write Utilities
// for writting to the encoder buffer

func (m *Encoder) write(b []byte) {
	m.buffer.Write(b)
}

func (m *Encoder) writeQuoted(b []byte) {
	m.buffer.WriteByte(m.quote)
	m.buffer.Write(b)
	m.buffer.WriteByte(m.quote)
}

func (m *Encoder) writeString(s string) {
	m.buffer.WriteString(s)
}

func (m *Encoder) writeQuotedString(s string) {
	m.buffer.WriteByte(m.quote)
	m.buffer.Write(EscapeString(s, m.quote, m.escape))
	m.buffer.WriteByte(m.quote)
}

// ----------------------------------------------------------------------------
// Encode Utilities
// methods for encoding to type

func (m *Encoder) Encode(a any) *Encoder {
	m.Reset()
	m.encode(t.ValueOf(a))
	return m
}

func (m *Encoder) encode(v t.Value, ancestry ...ancestor) {
	if v.IsNil() {
		m.write(m.Null)
		return
	}
	switch v.KindX() {
	case t.BOOL:
		m.encodeBool(v.Bool())
	case t.INT, t.INT8, t.INT16, t.INT32, t.INT64, t.UINT, t.UINT8, t.UINT16, t.UINT32, t.UINT64, t.UINTPTR, t.FLOAT32, t.FLOAT64, t.COMPLEX64, t.COMPLEX128:
		m.encodeNum(v)
	case t.ARRAY, t.SLICE:
		m.encodeSlice(v.Slice(), ancestry...)
	case t.FUNC:
		m.encodeFunc(v)
	case t.INTERFACE:
		m.encodeInterface(v, ancestry...)
	case t.MAP:
		m.encodeMap(v.Map(), ancestry...)
	case t.POINTER:
		m.encodePointer(v, ancestry...)
	case t.STRING:
		m.encodeString(*(*string)(v.Pointer()))
	case t.STRUCT:
		m.encodeStruct(v.Struct(), ancestry...)
	case t.UNSAFEPOINTER:
		m.encodeUnsafePointer(*(*unsafe.Pointer)(v.Pointer()))
	case t.TIME:
		m.encodeTime(*(*time.Time)(v.Pointer()))
	case t.UUID:
		m.encodeUuid(*(*uuid.UUID)(v.Pointer()))
	case t.BINARY:
		m.encodeString(v.Binary().String())
	case t.TYPE:
		m.encodeString((*t.Type)(v.Pointer()).String())
	default:
		panic("cannot encode type '" + v.Type().String() + "'")
	}
}

func (m *Encoder) encodeBool(b bool) {
	var bytes []byte
	if b {
		bytes = []byte("true")
	} else {
		bytes = []byte("false")
	}
	if m.QuotedBool {
		m.writeQuoted(bytes)
		return
	}
	m.write(bytes)
}

func (m *Encoder) encodeNum(v t.Value) {
	var bytes []byte
	switch v.Kind() {
	case t.INT:
		bytes = []byte(strconv.FormatInt(*(*int64)(v.Pointer()), 10))
	case t.INT8:
		bytes = []byte(strconv.FormatInt(int64(*(*int8)(v.Pointer())), 10))
	case t.INT16:
		bytes = []byte(strconv.FormatInt(int64(*(*int16)(v.Pointer())), 10))
	case t.INT32:
		bytes = []byte(strconv.FormatInt(int64(*(*int32)(v.Pointer())), 10))
	case t.INT64:
		bytes = []byte(strconv.FormatInt(*(*int64)(v.Pointer()), 10))
	case t.UINT:
		bytes = []byte(strconv.FormatUint(*(*uint64)(v.Pointer()), 10))
	case t.UINT8:
		bytes = []byte(strconv.FormatUint(uint64(*(*uint8)(v.Pointer())), 10))
	case t.UINT16:
		bytes = []byte(strconv.FormatUint(uint64(*(*uint16)(v.Pointer())), 10))
	case t.UINT32:
		bytes = []byte(strconv.FormatUint(uint64(*(*uint32)(v.Pointer())), 10))
	case t.UINT64:
		bytes = []byte(strconv.FormatUint(*(*uint64)(v.Pointer()), 10))
	case t.UINTPTR:
		bytes = []byte(strconv.FormatUint(*(*uint64)(v.Pointer()), 10))
	case t.FLOAT32:
		bytes = []byte(strconv.FormatFloat(float64(*(*float32)(v.Pointer())), 'f', -1, 64))
	case t.FLOAT64:
		bytes = []byte(strconv.FormatFloat(*(*float64)(v.Pointer()), 'f', -1, 64))
	case t.COMPLEX64:
		bytes = []byte(strconv.FormatComplex(complex128(*(*complex64)(v.Pointer())), 'f', -1, 128))
	case t.COMPLEX128:
		bytes = []byte(strconv.FormatComplex(*(*complex128)(v.Pointer()), 'f', -1, 128))
	default:
		panic("cannot encode type '" + v.Type().String() + "'")
	}
	if m.QuotedNum {
		m.writeQuoted(bytes)
		return
	}
	m.write(bytes)
}

func (m *Encoder) encodeFunc(v t.Value) {
	m.encodeString(v.Type().Name())
}

func (m *Encoder) encodeInterface(v t.Value, ancestry ...ancestor) {
	v = v.SetType()
	if v.Kind() != t.INTERFACE {
		m.encode(v, ancestry...)
		return
	}
	m.encodeString(fmt.Sprint(v.Interface()))
}

func (m *Encoder) encodePointer(v t.Value, ancestry ...ancestor) {
	ancestry = append([]ancestor{{v.Type(), v.Uintptr()}}, ancestry...)
	m.encode(v.Elem(), ancestry...)
}

func (m *Encoder) encodeMap(hm t.Map, ancestry ...ancestor) {
	if hm.Len() == 0 {
		m.encodeEmptyMap()
		return
	}
	delim, end, ancestry := m.encodeMapStart(hm.Value, ancestry)
	var j int
	hm.ForEach(func(k, v t.Value) (brake bool) {
		j = m.encodeElem(j, delim, k.Bytes(), v, ancestry)
		return
	})
	m.encodeEnd(end)
}

func (m *Encoder) encodeSlice(s t.Slice, ancestry ...ancestor) {
	if s.Len() == 0 {
		m.encodeEmptySlice()
		return
	}
	delim, end, ancestry := m.encodeSliceStart(s.Value, ancestry)
	var j int
	s.ForEach(func(i int, v t.Value) (brake bool) {
		j = m.encodeElem(j, delim, nil, v, ancestry)
		return
	})
	m.encodeEnd(end)
}

func (m *Encoder) encodeString(s string) {
	quoted := m.QuotedString
	if !quoted && m.QuotedSpecial {
		if ContainsSpecial(s) {
			quoted = true
		}
	}
	if quoted {
		m.writeQuotedString(s)
		return
	}
	m.writeString(s)
}

func (m *Encoder) encodeStruct(s t.Struct, ancestry ...ancestor) {
	if m.encodetStructByMethod(s) {
		return
	}
	if s.Len() == 0 {
		m.encodeEmptyMap()
		return
	}
	delim, end, ancestry := m.encodeMapStart(s.Value, ancestry)
	m.encodeStructTag(s)
	j, k, has, keys := 0, "", m.hasTag[s.Type()], m.tagKeys[s.Type()]
	s.ForEach(func(i int, f *t.FieldType, v t.Value) (brake bool) {
		if has {
			k = keys[i]
		} else {
			k = f.Name()
		}
		j = m.encodeElem(j, delim, []byte(k), v, ancestry)
		return
	})
	m.encodeEnd(end)
}

func (m *Encoder) encodetStructByMethod(s t.Struct) bool {
	if !m.EncodeMethods {
		return false
	}
	if m.methods != nil {
		if index, ok := m.methods[s.Type()]; ok {
			if index != -1 {
				m.write([]byte(s.Method(index).Call(nil)[0].String()))
				return true
			}
			return false
		}
	} else {
		m.methods = map[*t.Type]int{s.Type(): -1}
	}
	n := strings.ToUpper(m.Type[:1]) + m.Type[1:]
	methods := []string{n, "Encode" + n}
	for _, name := range methods {
		meth, exists := s.Type().Reflect().MethodByName(name)
		if exists {
			in, out := meth.Type.NumIn(), meth.Type.NumOut()
			if in == 1 && out > 0 {
				if k := t.FromReflectType(meth.Type.Out(0)).KindX(); k == t.STRING || k == t.BINARY {
					m.methods[s.Type()] = meth.Index
					m.write([]byte(s.Reflect().Method(meth.Index).Call(nil)[0].String()))
					return true
				}
			}
		}
	}
	return false
}

func (m *Encoder) encodeStructTag(s t.Struct) {
	if m.hasTag == nil {
		vals, has := s.Type().TagValues(m.Type)
		m.hasTag = map[*t.Type]bool{s.Type(): has}
		m.tagKeys = map[*t.Type][]string{s.Type(): vals}
	} else if _, ok := m.hasTag[s.Type()]; !ok {
		m.tagKeys[s.Type()], m.hasTag[s.Type()] = s.Type().TagValues(m.Type)
	}
}

func (m *Encoder) encodeUnsafePointer(p unsafe.Pointer) {
	m.encodeString(fmt.Sprintf("%p", p))
}

func (m *Encoder) encodeTime(t time.Time) {
	m.encodeString(t.String())
}

func (m *Encoder) encodeUuid(u uuid.UUID) {
	m.encodeString(u.String())
}

func (m *Encoder) encodeSliceComponents(v t.Value, ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	if !m.Format {
		return m.SliceStart, m.ValEnd, m.SliceEnd
	}
	path, hasDataElem := m.ancestryPath(v, ancestry)
	if parts, ok := m.sliceParts[path]; ok {
		return parts[0], parts[1], parts[2]
	}
	switch {
	case m.CascadeOnlyDeep && !hasDataElem:
		start, delim, end = m.InlineSyntax.SliceStart, m.ivalEnd, m.InlineSyntax.SliceEnd
	case m.hasBrackets:
		start, delim, end = m.formattedSliceComponents()
	default:
		start, delim, end = m.bracketlessSliceComponents(ancestry)
	}
	m.sliceParts[path] = [3][]byte{start, delim, end}
	return
}

func (m *Encoder) formattedSliceComponents() (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	start = append(append(append(append(m.SliceStart, m.LineBreak...), in...), m.Indent...), m.SliceItem...)
	delim = append(append(append(append(m.valEnd, m.LineBreak...), in...), m.Indent...), m.SliceItem...)
	end = append(append(m.LineBreak, in...), m.SliceEnd...)
	return
}

func (m *Encoder) bracketlessSliceComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	switch m.itemOf(ancestry) {
	case t.MAP:
		start = append(append(append(m.LineBreak, in...), m.Indent...), m.SliceItem...)
		delim = append(m.ValEnd, start...)
	case t.SLICE:
		start = m.SliceItem
		delim = append(append(append(append(append(m.ValEnd, m.LineBreak...), in...), m.Indent...), m.Indent...), m.SliceItem...)
	default:
		start = m.SliceItem
		delim = append(append(append(m.ValEnd, m.LineBreak...), in...), m.SliceItem...)
	}
	return
}

func (m *Encoder) encodeMapComponents(v t.Value, ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	if !m.Format {
		return m.MapStart, m.ValEnd, m.MapEnd
	}
	path, hasDataElem := m.ancestryPath(v, ancestry)
	if parts, ok := m.mapParts[path]; ok {
		return parts[0], parts[1], parts[2]
	}
	switch {
	case m.Format && m.CascadeOnlyDeep && !hasDataElem:
		start, delim, end = m.InlineSyntax.MapStart, m.ivalEnd, m.InlineSyntax.MapEnd
	case m.hasBrackets:
		start, delim, end = m.formattedMapComponents()
	default:
		start, delim, end = m.bracketlessMapComponents(ancestry)
	}
	m.mapParts[path] = [3][]byte{start, delim, end}
	return
}

func (m *Encoder) formattedMapComponents() (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	start = append(append(append(m.MapStart, m.LineBreak...), in...), m.Indent...)
	delim = append(append(append(m.valEnd, m.LineBreak...), in...), m.Indent...)
	end = append(append(m.LineBreak, in...), m.MapEnd...)
	return
}

func (m *Encoder) bracketlessMapComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	switch m.itemOf(ancestry) {
	case t.MAP:
		start = append(append(m.LineBreak, in...), m.Indent...)
		delim = append(append(append(m.ValEnd, m.LineBreak...), in...), m.Indent...)
	case t.SLICE:
		delim = append(append(append(m.ValEnd, m.LineBreak...), in...), m.Indent...)
	default:
		delim = append(append(m.ValEnd, m.LineBreak...), in...)
	}
	return
}

func (m *Encoder) itemOf(ancestry []ancestor) byte {
	k := m.encodeNonPtrParent(ancestry, 0)
	if k == t.MAP || k == t.STRUCT {
		return t.MAP
	}
	if (k == t.SLICE || k == t.ARRAY) && m.SliceItem != nil {
		return t.SLICE
	}
	return t.INVALID
}

func (m *Encoder) encodeNonPtrParent(ancestry []ancestor, pos int) byte {
	if len(ancestry) > pos {
		if k := ancestry[pos].typ.Kind(); k != t.POINTER {
			return k
		}
		return m.encodeNonPtrParent(ancestry, pos+1)
	}
	return t.INVALID
}

func (m *Encoder) encodeEmptySlice() {
	if !m.hasBrackets {
		m.write(m.Null)
		return
	}
	m.write(m.SliceStart)
	m.write(m.SliceEnd)
}

func (m *Encoder) encodeEmptyMap() {
	if !m.hasBrackets {
		m.write(m.Null)
		return
	}
	m.write(m.MapStart)
	m.write(m.MapEnd)
}

func (m *Encoder) encodeSliceStart(v t.Value, a []ancestor) (delim []byte, end []byte, ancestry []ancestor) {
	start, delim, end := m.encodeSliceComponents(v, a)
	m.write(start)
	m.IncDepth()
	return delim, end, append([]ancestor{{v.Type(), v.Uintptr()}}, a...)
}

func (m *Encoder) encodeMapStart(v t.Value, a []ancestor) (delim []byte, end []byte, ancestry []ancestor) {
	start, delim, end := m.encodeMapComponents(v, a)
	m.write(start)
	m.IncDepth()
	return delim, end, append([]ancestor{{v.Type(), v.Uintptr()}}, a...)
}

func (m *Encoder) encodeElem(i int, delim, k []byte, v t.Value, ancestry []ancestor) int {
	v = v.SetType()
	if i == 0 {
		delim = nil
	}
	if v.IsZero() {
		if !m.ExcludeZeros {
			i++
			if v.IsNil() {
				m.bufferElem(delim, k, m.Null)
			} else {
				m.bufferElem(delim, k, nil)
				m.encode(v)
			}
			return i
		}
		return i
	}
	if v.Type().IsData() {
		b, recursive := m.recursiveValue(v, ancestry)
		if recursive {
			if b != nil {
				i++
				m.bufferElem(delim, k, nil)
				m.encodeString(string(b))
			}
			return i
		}
	}
	i++
	m.bufferElem(delim, k, nil)
	m.encode(v, ancestry...)
	return i
}

func (m *Encoder) bufferElem(del, key, val []byte) {
	m.write(del)
	if key != nil {
		if m.QuotedKey {
			m.writeQuoted(key)
		} else {
			m.write(key)
		}
		m.write(m.keyEnd)
	}
	m.write(val)
}

func (m *Encoder) encodeEnd(end []byte) {
	m.write(end)
	m.decDepth()
}

func (m *Encoder) IncDepth() {
	m.curDepth++
	m.setIndent()
}

func (m *Encoder) decDepth() {
	m.curDepth--
	m.setIndent()
}

func (m *Encoder) setIndent() {
	if !m.hasBrackets && m.curDepth > 0 {
		m.curIndent = m.curDepth - 1
	} else {
		m.curIndent = m.curDepth
	}
}

func ContainsSpecial(s string) bool {
	for _, c := range s {
		if IsSpecialChar(byte(c)) {
			return true
		}
	}
	return false
}

func IsSpecialChar(b byte) bool {
	return b < 0x30 || (b > 0x3a && b < 0x41) || (b > 0x5a && b < 0x61) || b > 0x7a
}

func (m *Encoder) recursiveValue(v t.Value, ancestry []ancestor) (bytes []byte, is bool) {
	for _, a := range ancestry {
		if a.pointer == v.Uintptr() && a.typ == v.Type() {
			is = true
			if v.Kind() == t.STRUCT && m.RecursiveName {
				if m, ok := v.Type().Reflect().MethodByName("Name"); ok {
					bytes = []byte(v.Reflect().Method(m.Index).Call([]reflect.Value{})[0].String())
				} else if m, ok := v.Type().Reflect().MethodByName("String"); ok {
					bytes = []byte(v.Reflect().Method(m.Index).Call([]reflect.Value{})[0].String())
				} else {
					bytes = []byte(v.Type().NameShort())
				}
			}
			return
		}
	}
	if v.Kind() == t.POINTER {
		return m.recursiveValue(v.Elem(), ancestry)
	}
	return nil, false
}

func (m *Encoder) ancestryPath(v t.Value, ancestry []ancestor) (path string, hasDataElem bool) {
	elKind := m.dataElemKind(v)
	if elKind != t.INVALID {
		hasDataElem = true
	}
	path += m.ancestryPathVal(elKind) + m.ancestryPathVal(v.Kind())
	for _, a := range ancestry {
		if k := a.typ.Kind(); k != t.POINTER {
			path += m.ancestryPathVal(k)
		}
	}
	return
}

func (m *Encoder) ancestryPathVal(k byte) string {
	switch k {
	case t.MAP, t.STRUCT:
		return ":Map"
	case t.SLICE, t.ARRAY:
		return ":Slice"
	case t.INTERFACE:
		return ":Interface"
	default:
		return ":Value"
	}
}

func (m *Encoder) dataElemKind(v t.Value) (kind byte) {
	f := func(i int, e t.Value) (brake bool) {
		if i == 0 {
			kind = m.encodeKind(e.SetType().Type())
			return
		} else if m.encodeKind(e.SetType().Type()) != kind {
			kind = t.INTERFACE
			return true
		}
		return
	}
	switch v.Kind() {
	case t.ARRAY, t.SLICE:
		kind = m.encodeKind(v.Type().Elem())
		if kind == t.INTERFACE {
			v.Slice().ForEach(f)
		}
		return
	case t.INTERFACE:
		if v := v.SetType(); v.Kind() == t.INTERFACE {
			return t.INTERFACE
		}
		return m.dataElemKind(v)
	case t.MAP:
		if m.encodeKind(v.Type().Elem()) == t.INTERFACE {
			i := 0
			v.Map().ForEach(func(_, v t.Value) (brake bool) {
				brake = f(i, v)
				i++
				return
			})
		}
		return
	case t.POINTER:
		return m.dataElemKind(v.Elem())
	case t.STRUCT:
		kind = v.Type().Field(0).Type().Kind()
		v.Struct().ForEach(func(i int, _ *t.FieldType, v t.Value) (brake bool) {
			return f(i, v)
		})
		return
	default:
		return t.INVALID
	}
}

func (m *Encoder) encodeKind(typ *t.Type) byte {
	switch typ.DeepPtrElem().Kind() {
	case t.ARRAY, t.SLICE:
		return t.SLICE
	case t.MAP, t.STRUCT:
		return t.MAP
	case t.INTERFACE:
		return t.INTERFACE
	default:
		return t.INVALID
	}
}

// ----------------------------------------------------------------------------
// Decode Utilities
// methods for decoding from type

func (m *Encoder) decodeError(err string) {
	var start, mid, end int
	switch {
	case m.cursor == 0:
		start = 0
		mid = 0
		end = int(math.Min(float64(m.Len()-1), 25))
	case m.cursor >= m.Len():
		start = int(math.Max(0, float64(m.Len()-1-25)))
		mid = m.Len()
		end = m.Len()
	default:
		start = int(math.Max(0, float64(m.cursor-25)))
		mid = m.cursor
		end = int(math.Min(float64(m.Len()-1), float64(m.cursor+25)))
	}
	panic("decode error at position " + strconv.Itoa(m.cursor) + " of " + strconv.Itoa(m.Len()) + "\n" +
		"tried to decode:\n" + string(m.Buffer()[start:mid]) + "<ERROR>" + string(m.Buffer()[mid:end]) + "\n" +
		"error: " + err)
}

func (m *Encoder) Decode(bytes ...[]byte) *Encoder {
	m.ResetCursor()
	if len(bytes) > 0 {
		m.buffer.Set(bytes[0])
	}
	m.value = nil
	var slice []any
	var hmap map[string]any
	var value any
	for m.cursor < m.Len() {
		slice, hmap = m.decodeObject()
		if slice != nil {
			m.value = slice
			m.ResetCursor()
			return m
		}
		if hmap != nil {
			m.value = hmap
			m.ResetCursor()
			return m
		}
		value = m.decodeItem(nil)
		if value != any(nil) {
			m.value = value
			m.ResetCursor()
			return m
		}
		m.Inc()
	}
	m.ResetCursor()
	return m
}

func (m *Encoder) decodeObject(ancestry ...ancestor) (slice []any, hmap map[string]any) {
	if delim, end, isSlice := m.decodeSliceStart(ancestry); isSlice {
		return m.decodeSlice(delim, end, ancestry...), nil
	}
	if delim, end, isMap := m.decodeMapStart(ancestry); isMap {
		return nil, m.decodeMap(delim, end, ancestry...)
	}
	return nil, nil
}

func (m *Encoder) decodeSlice(delim, end []byte, ancestry ...ancestor) (slice []any) {
	ancestry = append([]ancestor{{sliceType, 0}}, ancestry...)
	for m.cursor < m.Len() {
		slice = append(slice, m.decodeItem([][]byte{delim, end}, ancestry...))
		m.decodeNonData()
		if m.isMatch(delim) {
			m.Inc(len(delim))
			continue
		}
		if m.isMatch(end) || end == nil {
			m.Inc(len(end))
			m.decDepth()
			break
		}
		m.decodeError("failed to find end of slice element")
	}
	return
}

func (m *Encoder) decodeMap(delim, end []byte, ancestry ...ancestor) map[string]any {
	ancestry = append([]ancestor{{mapType, 0}}, ancestry...)
	hmap := map[string]any{}
	for m.cursor < m.Len() {
		hmap[m.decodeKey()] = m.decodeItem([][]byte{delim, end}, ancestry...)
		m.decodeNonData()
		if m.isMatch(delim) {
			m.Inc(len(delim))
			continue
		} else if m.isMatch(end) || end == nil {
			m.Inc(len(end))
			m.decDepth()
			break
		}
		m.decodeError("failed to find end of map element")
	}
	return hmap
}

func (m *Encoder) decodeItem(endings [][]byte, ancestry ...ancestor) any {
	m.decodeNonData()
	switch {
	case m.isQuote():
		return m.decodeQuote()
	case m.isNull():
		return m.decodeNull()
	default:
		slice, hmap := m.decodeObject(ancestry...)
		switch {
		case slice != nil:
			return slice
		case hmap != nil:
			return hmap
		default:
			return m.decodeAny(endings...)
		}
	}
}

func (m *Encoder) decodeKey() (key string) {
	m.decodeNonData()
	if m.isQuote() {
		key = m.decodeQuote()
	}
	s := m.cursor
	for m.cursor < m.Len() {
		if m.isKeyEnd() {
			m.Inc()
			break
		}
		m.Inc()
	}
	if key == "" {
		key = string(m.Buffer()[s : m.cursor-1])
	}
	return
}

func (m *Encoder) decodeAny(end ...[]byte) any {
	s := m.cursor
	for m.cursor < m.Len() {
		if m.isMatch(end...) || m.isMatch(m.LineBreak) {
			break
		}
		m.Inc()
	}
	a := strings.Trim(string(m.Buffer()[s:m.cursor]), string(m.Space))
	if a == "" {
		return nil
	}
	if m.DecodeTyped {
		if a == "true" {
			return true
		}
		if a == "false" {
			return false
		}
		if a == string(m.Null) {
			return nil
		}
		if i, e := strconv.ParseInt(a, 10, 64); e == nil {
			return int(i)
		}
		if f, e := strconv.ParseFloat(a, 64); e == nil {
			return f
		}
	}
	return a
}

func (m *Encoder) decodeQuote() string {
	q := m.Buffer()[m.cursor]
	m.Inc()
	s := m.cursor
	for m.cursor < m.Len() {
		if m.isEscape() {
			m.Inc(2)
			continue
		}
		if m.ByteIs(q) {
			m.Inc()
			break
		}
		m.Inc()
	}
	return string(m.Buffer()[s : m.cursor-1])
}

func (m *Encoder) decodeNull() any {
	m.Inc(len(m.Null))
	return nil
}

func (m *Encoder) decodeSpace() string {
	s := m.cursor
	for m.cursor < m.Len() {
		if !m.isSpace() {
			break
		}
		m.Inc()
	}
	return string(m.Buffer()[s : m.cursor-1])
}

func (m *Encoder) decodeCommentBlock() []byte {
	if m.isBlockCommentStart() {
		s := m.cursor
		for m.cursor < m.Len() {
			if m.isBlockCommentEnd() {
				m.Inc(len(m.BlockCommentEnd))
				break
			}
			m.Inc()
		}
		return m.Buffer()[s : m.cursor+len(m.BlockCommentEnd)-1]
	}
	return nil
}

func (m *Encoder) decodeInlineComment() []byte {
	if m.isLineCommentStart() {
		s := m.cursor
		for m.cursor < m.Len() {
			if m.isLineCommentEnd() {
				m.Inc(len(m.LineCommentEnd))
				break
			}
			m.Inc()
		}
		return m.Buffer()[s : m.cursor+len(m.LineCommentEnd)-1]
	}
	return nil
}

func (m *Encoder) decodeNonData() []byte {
	data, s := false, m.cursor
	for m.cursor < m.Len() && !data {
		switch {
		case m.isBlockCommentStart():
			m.decodeCommentBlock()
		case m.isLineCommentStart():
			m.decodeInlineComment()
		case m.isSpace():
			m.decodeSpace()
		default:
			data = true
		}
	}
	return m.Buffer()[s:m.cursor]
}

func (m *Encoder) DecodeTo(stop []byte) []byte {
	s := m.cursor
	e := stop[0]
	for m.cursor < m.Len() {
		if m.ByteIs(e) {
			if MatchBytes(m.Buffer()[m.cursor:m.cursor+len(stop)], stop) {
				m.Inc(len(stop))
				break
			}
		}
	}
	return m.Buffer()[s:m.cursor]
}

func (m *Encoder) decodeSliceStart(ancestry []ancestor) (delim, end []byte, is bool) {
	if m.hasBrackets {
		if m.isMatch(m.SliceStart) {
			m.Inc(len(m.SliceStart))
			m.IncDepth()
			return m.ValEnd, m.SliceEnd, true
		}
	} else {
		var start []byte
		if start, delim, end = m.bracketlessSliceComponents(ancestry); m.isMatch(start) {
			m.Inc(len(start))
			m.IncDepth()
			return delim, end, true
		}
	}
	if m.Format {
		if m.isMatch(m.InlineSyntax.SliceStart) {
			m.Inc(len(m.InlineSyntax.SliceStart))
			m.IncDepth()
			return m.InlineSyntax.ValEnd, m.InlineSyntax.SliceEnd, true
		}
	}
	return nil, nil, false
}

func (m *Encoder) decodeMapStart(ancestry []ancestor) (delim, end []byte, is bool) {
	if m.hasBrackets {
		if m.isMatch(m.MapStart) {
			m.Inc(len(m.MapStart))
			m.IncDepth()
			return m.ValEnd, m.MapEnd, true
		}
	} else {
		var start []byte
		if start, delim, end = m.bracketlessMapComponents(ancestry); start != nil {
			if m.isMatch(start) {
				m.Inc(len(start))
				m.IncDepth()
				return delim, end, true
			}
		} else if m.isKey() {
			m.IncDepth()
			return delim, end, true
		}
	}
	if m.Format {
		if m.isMatch(m.InlineSyntax.MapStart) {
			m.Inc(len(m.InlineSyntax.MapStart))
			m.IncDepth()
			return m.InlineSyntax.ValEnd, m.InlineSyntax.MapEnd, true
		}
	}
	return nil, nil, false
}

func (m *Encoder) Inc(i ...int) {
	if i == nil {
		m.cursor++
		return
	}
	m.cursor += i[0]
}

func (m *Encoder) Byte() byte {
	if m.Len() == 0 {
		return 0
	}
	return m.Buffer()[m.cursor]
}

func (m *Encoder) ByteIs(b byte) bool {
	return m.Byte() == b
}

func (m *Encoder) isSpace() bool {
	return InBytes(m.Buffer()[m.cursor], m.Space)
}

func (m *Encoder) isQuote() bool {
	return InBytes(m.Buffer()[m.cursor], m.Quote)
}

func (m *Encoder) isEscape() bool {
	return InBytes(m.Buffer()[m.cursor], m.Escape)
}

func (m *Encoder) isNull() bool {
	return m.isMatch(m.Null)
}

func (m *Encoder) isKey() bool {
	i := m.cursor
	l := len(m.KeyEnd)
	for i < m.Len() {
		if m.Buffer()[i] == m.LineBreak[0] {
			return false
		}
		if MatchBytes(m.Buffer()[i:i+l], m.KeyEnd) {
			return true
		}
		i++
	}
	return false
}

func (m *Encoder) isKeyEnd() bool {
	return InBytes(m.Buffer()[m.cursor], m.KeyEnd)
}

func (m *Encoder) isBlockCommentStart() bool {
	return m.isMatch(m.BlockCommentStart)
}

func (m *Encoder) isBlockCommentEnd() bool {
	return m.isMatch(m.BlockCommentEnd)
}
func (m *Encoder) isLineCommentStart() bool {
	return m.isMatch(m.LineCommentStart)
}

func (m *Encoder) isLineCommentEnd() bool {
	return m.isMatch(m.LineCommentEnd)
}

func (m *Encoder) isMatch(b ...[]byte) bool {
	for _, s := range b {
		if m.Len()-m.cursor < len(s) {
			continue
		}
		if MatchBytes(m.Buffer()[m.cursor:m.cursor+len(s)], s) {
			return true
		}
	}
	return false
}

func MatchBytes(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}

func InBytes(a byte, b []byte) bool {
	if b == nil {
		return false
	}
	for _, s := range b {
		if s == a {
			return true
		}
	}
	return false
}

func EscapeBytes(b []byte, escape byte, chars ...byte) (escaped []byte) {
	escaped = make([]byte, 0, len(b))
	for _, c := range b {
		if InBytes(c, chars) {
			escaped = append(escaped, escape)
		}
		escaped = append(escaped, c)
	}
	return escaped
}

func EscapeString(s string, escape byte, chars ...byte) (escaped []byte) {
	escaped = make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if InBytes(s[i], chars) {
			escaped = append(escaped, escape)
		}
		escaped = append(escaped, s[i])
	}
	return escaped
}
