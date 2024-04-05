// Copyright 2023 james dotter. All rights reserved.typVal
// Use of this source code is governed by a
// license that can be found in the gotype LICENSE file.

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

	t "github.com/jcdotter/go/typ"
	"github.com/jcdotter/go/uuid"
)

// ------------------------------------------------------------ /
// Speed:
// 1. Pass structure parts to children and append indents, rather than using ancestry
// 2. encoding pkg does not handle recursive structures
// ------------------------------------------------------------ /

// ------------------------------------------------------------ /
// MARSHALER IMPLEMENTATION
// a generic marshaller for serializing and deserializing
// golang values to and from []byte in a specific format
// ------------------------------------------------------------ /

type Marshaler struct {
	// marshaler state
	cursor      int    // the current position in the buffer
	curDepth    int    // the current depth of the data structure
	curIndent   int    // the current indentation level
	hasBrackets bool   // true when marshalling with brackets around data objects
	value       any    // the value being marshalled
	buffer      []byte // the buffer being marshalled to
	len         int    // the length of the buffer
	availBuf    int    // the available buffer space
	// marshaling syntax
	Type              string // the type of marshaller. json, yaml, etc.
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
	// marshaling flags
	Format           bool // when true, marshal with formatting, indentation, and line breaks
	FormatWithSpaces bool // when true, marshal with space between keys and values
	CascadeOnlyDeep  bool // when true, marshal single-depth slices and maps with inline syntax
	QuotedKey        bool // when true, marshal map keys with quotes
	QuotedString     bool // when true, marshal strings with quotes
	QuotedSpecial    bool // when true, marshal strings with quotes if they contain special characters
	QuotedNum        bool // when true, marshal numbers with quotes
	QuotedBool       bool // when true, marshal bools with quotes
	QuotedNull       bool // when true, marshal null with quotes
	RecursiveName    bool // when true, include name, string or type of recursive value in marshalling, otherwise, exclude all recursion
	UnmarshalTyped   bool // when true, unmarshal to typed values (int, float64, bool, string) instead of just strings
	MarshalMethods   bool // when true, marshal structs with a Marshal method by calling the method
	ExcludeZeros     bool // when true, exclude zero and nil values from marshalling
	// marshaler cache
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

// ------------------------------------------------------------ /
// PRESET MARSHALERS
// JSON, YAML...
// ------------------------------------------------------------ /

var (
	JsonMarshaler = &Marshaler{
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
	YamlMarshaler = &Marshaler{
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

// ------------------------------------------------------------ /
// Init Utilities
// methods for setting up the marshaler
// ------------------------------------------------------------ /

func init() {
	JsonMarshaler.Init()
	YamlMarshaler.Init()
}

func (m *Marshaler) Init() {
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
		panic("cannot marshal without brackets or formatting: unable to determine data structure")
	}
	if m.SliceItem != nil && m.hasBrackets {
		panic("slice item reserved for bracketless marshalling")
	}
	if !m.hasBrackets && InBytes('\n', m.Space) {
		panic("cannot marshal without brackets when line breaks in space characters")
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

func (m *Marshaler) initFormat() {
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

func (m *Marshaler) New() *Marshaler {
	n := *m
	if m.InlineSyntax != nil {
		s := *m.InlineSyntax
		n.InlineSyntax = &s
	}
	return &n
}

func (m *Marshaler) Reset() {
	m.availBuf = 10
	m.buffer = make([]byte, 0, m.availBuf)
	m.cursor = 0
	m.curIndent = 0
	m.len = 0
	m.value = nil
	m.sliceParts = map[string][3][]byte{}
	m.mapParts = map[string][3][]byte{}
	m.hasTag = nil
	m.tagKeys = nil
	m.methods = nil
}

func (m *Marshaler) ResetCursor() {
	m.cursor = 0
	m.curDepth = 0
	m.curIndent = 0
}

// ------------------------------------------------------------ /
// Getter Utilities
// for getting non-exported state values from the marshaler
// ------------------------------------------------------------ /

func (m *Marshaler) Buffer() []byte {
	return m.buffer
}

func (m *Marshaler) Cursor() int {
	return m.cursor
}

func (m *Marshaler) CurIndent() int {
	return m.curIndent
}

func (m *Marshaler) Len() int {
	return m.len
}

func (m *Marshaler) Value() any {
	return m.value
}

func (m Marshaler) Bytes() []byte {
	return m.buffer
}

func (m Marshaler) String() string {
	return string(m.buffer)
}

func (m Marshaler) Formatted(indent ...int) string {
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
		m.Marshal(m.value)
		m.Indent = oldIndent
		m.Format = oldFormat
	}
	m.LineBreak = oldBreak
	return m.String()
}

func (m Marshaler) Map() map[string]any {
	v := t.ValueOf(m.value)
	switch v.Kind() {
	case t.MAP:
		return m.value.(map[string]any)
	case t.SLICE:
		return v.Slice().Map()
	}
	return nil
}

func (m Marshaler) Slice() []any {
	v := t.ValueOf(m.value)
	switch v.Kind() {
	case t.MAP:
		return v.Map().Values()
	case t.SLICE:
		return m.value.([]any)
	}
	return nil
}

// ------------------------------------------------------------ /
// Marshal Utilities
// methods for marshaling to type
// ------------------------------------------------------------ /

func (m *Marshaler) Marshal(a any) *Marshaler {
	m.Reset()
	m.marshal(t.ValueOf(a))
	m.setLen()
	return m
}

func (m *Marshaler) marshal(v t.Value, ancestry ...ancestor) {
	if v.IsNil() {
		m.bufferBytes(m.Null)
		return
	}
	switch v.KindX() {
	case t.BOOL:
		m.marshalBool(v.Bool())
	case t.INT, t.INT8, t.INT16, t.INT32, t.INT64, t.UINT, t.UINT8, t.UINT16, t.UINT32, t.UINT64, t.UINTPTR, t.FLOAT32, t.FLOAT64, t.COMPLEX64, t.COMPLEX128:
		m.marshalNum(v)
	case t.ARRAY:
		m.marshalArray(v.Slice(), ancestry...)
	case t.FUNC:
		m.marshalFunc(v)
	case t.INTERFACE:
		m.marshalInterface(v, ancestry...)
	case t.MAP:
		m.marshalMap(v.Map(), ancestry...)
	case t.POINTER:
		m.marshalPointer(v, ancestry...)
	case t.SLICE:
		m.marshalSlice(v.Slice(), ancestry...)
	case t.STRING:
		m.marshalString(*(*string)(v.Pointer()))
	case t.STRUCT:
		m.marshalStruct(v.Struct(), ancestry...)
	case t.UNSAFEPOINTER:
		m.marshalUnsafePointer(*(*unsafe.Pointer)(v.Pointer()))
	case t.TIME:
		m.marshalTime(*(*time.Time)(v.Pointer()))
	case t.UUID:
		m.marshalUuid(*(*uuid.UUID)(v.Pointer()))
	case t.BINARY:
		m.marshalString(v.Binary().String())
	case t.TYPE:
		m.marshalString((*t.Type)(v.Pointer()).String())
	default:
		panic("cannot marshal type '" + v.Type().String() + "'")
	}
}

func (m *Marshaler) marshalBool(b bool) {
	var bytes []byte
	if b {
		bytes = []byte("true")
	} else {
		bytes = []byte("false")
	}
	if m.QuotedBool {
		m.SetBuffer(append(append(append(m.buffer, m.quote), bytes...), m.quote))
		return
	}
	m.bufferBytes(bytes)
}

func (m *Marshaler) marshalNum(v t.Value) {
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
		panic("cannot marshal type '" + v.Type().String() + "'")
	}
	if m.QuotedNum {
		m.SetBuffer(append(append(append(m.buffer, m.quote), bytes...), m.quote))
		return
	}
	m.bufferBytes(bytes)
}

func (m *Marshaler) marshalArray(a t.Slice, ancestry ...ancestor) {
	if a.Len() == 0 {
		m.marshalEmptySlice()
		return
	}
	delim, end, ancestry := m.marshalSliceStart(a.Value, ancestry)
	var j int
	a.ForEach(func(i int, v t.Value) (brake bool) {
		j = m.marshalElem(j, delim, nil, v, ancestry)
		return
	})
	m.marshalEnd(end)
}

func (m *Marshaler) marshalFunc(v t.Value) {
	m.bufferBytes([]byte(v.Type().Name()))
}

func (m *Marshaler) marshalInterface(v t.Value, ancestry ...ancestor) {
	v = v.SetType()
	if v.Kind() != t.INTERFACE {
		m.marshal(v, ancestry...)
		return
	}
	m.marshalString(fmt.Sprint(v.Interface()))
}

func (m *Marshaler) marshalPointer(v t.Value, ancestry ...ancestor) {
	ancestry = append([]ancestor{{v.Type(), v.Uintptr()}}, ancestry...)
	m.marshal(v.Elem(), ancestry...)
}

func (m *Marshaler) marshalMap(hm t.Map, ancestry ...ancestor) {
	if hm.Len() == 0 {
		m.marshalEmptyMap()
		return
	}
	delim, end, ancestry := m.marshalMapStart(hm.Value, ancestry)
	var j int
	hm.ForEach(func(k, v t.Value) (brake bool) {
		j = m.marshalElem(j, delim, k.Bytes(), v, ancestry)
		return
	})
	m.marshalEnd(end)
}

func (m *Marshaler) marshalSlice(s t.Slice, ancestry ...ancestor) {
	if s.Len() == 0 {
		m.marshalEmptySlice()
		return
	}
	delim, end, ancestry := m.marshalSliceStart(s.Value, ancestry)
	var j int
	s.ForEach(func(i int, v t.Value) (brake bool) {
		j = m.marshalElem(j, delim, nil, v, ancestry)
		return
	})
	m.marshalEnd(end)
}

func (m *Marshaler) marshalString(s string) {
	b := []byte(s)
	quoted := m.QuotedString
	if !quoted && m.QuotedSpecial {
		if ContainsSpecial(s) {
			quoted = true
		}
	}
	if quoted {
		m.SetBuffer(append(append(append(m.buffer, m.quote), EscapeBytes(b, m.quote, m.escape)...), m.quote))
		return
	}
	m.bufferBytes(b)
}

func (m *Marshaler) marshalStruct(s t.Struct, ancestry ...ancestor) {
	if m.marshaltStructByMethod(s) {
		return
	}
	if s.Len() == 0 {
		m.marshalEmptyMap()
		return
	}
	delim, end, ancestry := m.marshalMapStart(s.Value, ancestry)
	m.marshalStructTag(s)
	j, k, has, keys := 0, "", m.hasTag[s.Type()], m.tagKeys[s.Type()]
	s.ForEach(func(i int, f *t.FieldType, v t.Value) (brake bool) {
		if has {
			k = keys[i]
		} else {
			k = f.Name()
		}
		j = m.marshalElem(j, delim, []byte(k), v, ancestry)
		return
	})
	m.marshalEnd(end)
}

func (m *Marshaler) marshaltStructByMethod(s t.Struct) bool {
	if !m.MarshalMethods {
		return false
	}
	if m.methods != nil {
		if index, ok := m.methods[s.Type()]; ok {
			if index != -1 {
				m.bufferBytes([]byte(s.Method(index).Call(nil)[0].String()))
				return true
			}
			return false
		}
	} else {
		m.methods = map[*t.Type]int{s.Type(): -1}
	}
	n := strings.ToUpper(m.Type[:1]) + m.Type[1:]
	methods := []string{n, "Marshal" + n}
	for _, name := range methods {
		meth, exists := s.Type().Reflect().MethodByName(name)
		if exists {
			in, out := meth.Type.NumIn(), meth.Type.NumOut()
			if in == 1 && out > 0 {
				if k := t.FromReflectType(meth.Type.Out(0)).KindX(); k == t.STRING || k == t.BINARY {
					m.methods[s.Type()] = meth.Index
					m.bufferBytes([]byte(s.Reflect().Method(meth.Index).Call(nil)[0].String()))
					return true
				}
			}
		}
	}
	return false
}

func (m *Marshaler) marshalStructTag(s t.Struct) {
	if m.hasTag == nil {
		vals, has := s.Type().TagValues(m.Type)
		m.hasTag = map[*t.Type]bool{s.Type(): has}
		m.tagKeys = map[*t.Type][]string{s.Type(): vals}
	} else if _, ok := m.hasTag[s.Type()]; !ok {
		m.tagKeys[s.Type()], m.hasTag[s.Type()] = s.Type().TagValues(m.Type)
	}
}

func (m *Marshaler) marshalUnsafePointer(p unsafe.Pointer) {
	m.marshalString(fmt.Sprintf("%p", p))
}

func (m *Marshaler) marshalTime(t time.Time) {
	m.marshalString(t.String())
}

func (m *Marshaler) marshalUuid(u uuid.UUID) {
	m.marshalString(u.String())
}

func (m *Marshaler) marshalSliceComponents(v t.Value, ancestry []ancestor) (start []byte, delim []byte, end []byte) {
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
		start, delim, end = m.formattedSliceComponents(ancestry)
	default:
		start, delim, end = m.bracketlessSliceComponents(ancestry)
	}
	m.sliceParts[path] = [3][]byte{start, delim, end}
	return
}

func (m *Marshaler) formattedSliceComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	start = append(append(append(append(m.SliceStart, m.LineBreak...), in...), m.Indent...), m.SliceItem...)
	delim = append(append(append(append(m.valEnd, m.LineBreak...), in...), m.Indent...), m.SliceItem...)
	end = append(append(m.LineBreak, in...), m.SliceEnd...)
	return
}

func (m *Marshaler) bracketlessSliceComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
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

func (m *Marshaler) marshalMapComponents(v t.Value, ancestry []ancestor) (start []byte, delim []byte, end []byte) {
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
		start, delim, end = m.formattedMapComponents(ancestry)
	default:
		start, delim, end = m.bracketlessMapComponents(ancestry)
	}
	m.mapParts[path] = [3][]byte{start, delim, end}
	return
}

func (m *Marshaler) formattedMapComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
	in := bytes.Repeat(m.Indent, m.curIndent)
	start = append(append(append(m.MapStart, m.LineBreak...), in...), m.Indent...)
	delim = append(append(append(m.valEnd, m.LineBreak...), in...), m.Indent...)
	end = append(append(m.LineBreak, in...), m.MapEnd...)
	return
}

func (m *Marshaler) bracketlessMapComponents(ancestry []ancestor) (start []byte, delim []byte, end []byte) {
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

func (m *Marshaler) itemOf(ancestry []ancestor) byte {
	k := m.marshalNonPtrParent(ancestry, 0)
	if k == t.MAP || k == t.STRUCT {
		return t.MAP
	}
	if (k == t.SLICE || k == t.ARRAY) && m.SliceItem != nil {
		return t.SLICE
	}
	return t.INVALID
}

func (m *Marshaler) marshalNonPtrParent(ancestry []ancestor, pos int) byte {
	if len(ancestry) > pos {
		if k := ancestry[pos].typ.Kind(); k != t.POINTER {
			return k
		}
		return m.marshalNonPtrParent(ancestry, pos+1)
	}
	return t.INVALID
}

func (m *Marshaler) marshalEmptySlice() {
	if !m.hasBrackets {
		m.bufferBytes(m.Null)
		return
	}
	m.SetBuffer(append(append(m.buffer, m.SliceStart...), m.SliceEnd...))
}

func (m *Marshaler) marshalEmptyMap() {
	if !m.hasBrackets {
		m.bufferBytes(m.Null)
		return
	}
	m.SetBuffer(append(append(m.buffer, m.MapStart...), m.MapEnd...))
}

func (m *Marshaler) marshalSliceStart(v t.Value, a []ancestor) (delim []byte, end []byte, ancestry []ancestor) {
	start, delim, end := m.marshalSliceComponents(v, a)
	m.bufferBytes(start)
	m.IncDepth()
	return delim, end, append([]ancestor{{v.Type(), v.Uintptr()}}, a...)
}

func (m *Marshaler) marshalMapStart(v t.Value, a []ancestor) (delim []byte, end []byte, ancestry []ancestor) {
	start, delim, end := m.marshalMapComponents(v, a)
	m.bufferBytes(start)
	m.IncDepth()
	return delim, end, append([]ancestor{{v.Type(), v.Uintptr()}}, a...)
}

func (m *Marshaler) marshalElem(i int, delim, k []byte, v t.Value, ancestry []ancestor) int {
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
				m.marshal(v)
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
				m.marshalString(string(b))
			}
			return i
		}
	}
	i++
	m.bufferElem(delim, k, nil)
	m.marshal(v, ancestry...)
	return i
}

func (m *Marshaler) bufferElem(del, key, val []byte) {
	if key != nil {
		if m.QuotedKey {
			m.SetBuffer(append(append(append(append(append(append(m.buffer, del...), m.quote), key...), m.quote), m.keyEnd...), val...))
			return
		}
		m.SetBuffer(append(append(append(append(m.buffer, del...), key...), m.keyEnd...), val...))
		return
	}
	m.SetBuffer(append(append(m.buffer, del...), val...))
}

func (m *Marshaler) marshalEnd(end []byte) {
	m.bufferBytes(end)
	m.decDepth()
}

func (m *Marshaler) IncDepth() {
	m.curDepth++
	m.setIndent()
}

func (m *Marshaler) decDepth() {
	m.curDepth--
	m.setIndent()
}

func (m *Marshaler) setIndent() {
	if !m.hasBrackets && m.curDepth > 0 {
		m.curIndent = m.curDepth - 1
	} else {
		m.curIndent = m.curDepth
	}
}

func (m *Marshaler) bufferBytes(b []byte) {
	m.buffer = append(m.buffer, b...)
}

func (m *Marshaler) setLen() {
	m.len = len(m.buffer)
}

// first arg should be the buffer to append,
// or not, to replace the buffer
func (m *Marshaler) SetBuffer(b []byte) {
	m.buffer = b
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

func (m *Marshaler) recursiveValue(v t.Value, ancestry []ancestor) (bytes []byte, is bool) {
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

func (m *Marshaler) ancestryPath(v t.Value, ancestry []ancestor) (path string, hasDataElem bool) {
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

func (m *Marshaler) ancestryPathVal(k byte) string {
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

func (m *Marshaler) dataElemKind(v t.Value) (kind byte) {
	f := func(i int, k string, e t.Value) (brake bool) {
		if i == 0 {
			kind = m.marshalKind(e.SetType().Type())
			return
		}
		if m.marshalKind(e.SetType().Type()) != kind {
			kind = t.INTERFACE
			return true
		}
		return
	}
	switch v.Kind() {
	case t.ARRAY:
		kind = m.marshalKind((*arrayType)(unsafe.Pointer(v.Type())).elem)
		if kind == t.INTERFACE {
			v.Slice().ForEach(f)
		}
		return
	case Interface:
		if v := v.SetType(); v.Kind() == Interface {
			return Interface
		}
		return m.dataElemKind(v)
	case Map:
		kind = m.marshalKind((*mapType)(unsafe.Pointer(v.typ)).elem)
		if kind == Interface {
			(MAP)(v).ForEach(f)
		}
		return
	case Pointer:
		return m.dataElemKind(v.Elem())
	case Slice:
		kind = m.marshalKind((*sliceType)(unsafe.Pointer(v.typ)).elem)
		if kind == Interface {
			(SLICE)(v).ForEach(f)
		}
		return
	case Struct:
		kind = m.marshalKind((*structType)(unsafe.Pointer(v.typ)).fields[0].typ)
		(STRUCT)(v).ForEach(f)
		return
	default:
		return Invalid
	}
}

func (m *Marshaler) marshalKind(typ *t.Type) byte {
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

// ------------------------------------------------------------ /
// Unmarshal Utilities
// methods for unmarshaling from type
// ------------------------------------------------------------ /

func (m *Marshaler) unmarshalError(err string) {
	var start, mid, end int
	switch {
	case m.cursor == 0:
		start = 0
		mid = 0
		end = int(math.Min(float64(m.len-1), 25))
	case m.cursor >= m.len:
		start = int(math.Max(0, float64(m.len-1-25)))
		mid = m.len
		end = m.len
	default:
		start = int(math.Max(0, float64(m.cursor-25)))
		mid = m.cursor
		end = int(math.Min(float64(m.len-1), float64(m.cursor+25)))
	}
	panic("unmarshal error at position " + strconv.Itoa(m.cursor) + " of " + strconv.Itoa(m.len) + "\n" +
		"tried to unmarshal:\n" + string(m.buffer[start:mid]) + "<ERROR>" + string(m.buffer[mid:end]) + "\n" +
		"error: " + err)
}

func (m *Marshaler) Unmarshal(bytes ...[]byte) *Marshaler {
	m.ResetCursor()
	if len(bytes) > 0 {
		m.buffer = bytes[0]
		m.len = len(m.buffer)
	}
	m.value = nil
	var slice []any
	var hmap map[string]any
	var value any
	for m.cursor < m.len {
		slice, hmap = m.unmarshalObject()
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
		value = m.unmarshalItem(nil)
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

func (m *Marshaler) unmarshalObject(ancestry ...ancestor) (slice []any, hmap map[string]any) {
	if delim, end, isSlice := m.unmarshalSliceStart(ancestry); isSlice {
		return m.unmarshalSlice(delim, end, ancestry...), nil
	}
	if delim, end, isMap := m.unmarshalMapStart(ancestry); isMap {
		return nil, m.unmarshalMap(delim, end, ancestry...)
	}
	return nil, nil
}

func (m *Marshaler) unmarshalSlice(delim, end []byte, ancestry ...ancestor) (slice []any) {
	ancestry = append([]ancestor{{&TYPE{kind: 23}, 0}}, ancestry...)
	for m.cursor < m.len {
		slice = append(slice, m.unmarshalItem([][]byte{delim, end}, ancestry...))
		m.unmarshalNonData()
		if m.isMatch(delim) {
			m.Inc(len(delim))
			continue
		}
		if m.isMatch(end) || end == nil {
			m.Inc(len(end))
			m.decDepth()
			break
		}
		m.unmarshalError("failed to find end of slice element")
	}
	return
}

func (m *Marshaler) unmarshalMap(delim, end []byte, ancestry ...ancestor) map[string]any {
	ancestry = append([]ancestor{{&TYPE{kind: 53}, 0}}, ancestry...)
	hmap := map[string]any{}
	for m.cursor < m.len {
		hmap[m.unmarshalKey()] = m.unmarshalItem([][]byte{delim, end}, ancestry...)
		m.unmarshalNonData()
		if m.isMatch(delim) {
			m.Inc(len(delim))
			continue
		} else if m.isMatch(end) || end == nil {
			m.Inc(len(end))
			m.decDepth()
			break
		}
		m.unmarshalError("failed to find end of map element")
	}
	return hmap
}

func (m *Marshaler) unmarshalItem(endings [][]byte, ancestry ...ancestor) any {
	m.unmarshalNonData()
	switch {
	case m.isQuote():
		return m.unmarshalQuote()
	case m.isNull():
		return m.unmarshalNull()
	default:
		slice, hmap := m.unmarshalObject(ancestry...)
		switch {
		case slice != nil:
			return slice
		case hmap != nil:
			return hmap
		default:
			return m.unmarshalAny(endings...)
		}
	}
}

func (m *Marshaler) unmarshalKey() (key string) {
	m.unmarshalNonData()
	if m.isQuote() {
		key = m.unmarshalQuote()
	}
	s := m.cursor
	for m.cursor < m.len {
		if m.isKeyEnd() {
			m.Inc()
			break
		}
		m.Inc()
	}
	if key == "" {
		key = string(m.buffer[s : m.cursor-1])
	}
	return
}

func (m *Marshaler) unmarshalAny(end ...[]byte) any {
	s := m.cursor
	for m.cursor < m.len {
		if m.isMatch(end...) || m.isMatch(m.LineBreak) {
			break
		}
		m.Inc()
	}
	a := STRING(m.buffer[s:m.cursor]).Trim(string(m.Space))
	if a == "" {
		return nil
	}
	if m.UnmarshalTyped {
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

func (m *Marshaler) unmarshalQuote() string {
	q := m.buffer[m.cursor]
	m.Inc()
	s := m.cursor
	for m.cursor < m.len {
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
	return string(m.buffer[s : m.cursor-1])
}

func (m *Marshaler) unmarshalNull() any {
	m.Inc(len(m.Null))
	return nil
}

func (m *Marshaler) unmarshalSpace() string {
	s := m.cursor
	for m.cursor < m.len {
		if !m.isSpace() {
			break
		}
		m.Inc()
	}
	return string(m.buffer[s : m.cursor-1])
}

func (m *Marshaler) unmarshalCommentBlock() []byte {
	if m.isBlockCommentStart() {
		s := m.cursor
		for m.cursor < m.len {
			if m.isBlockCommentEnd() {
				m.Inc(len(m.BlockCommentEnd))
				break
			}
			m.Inc()
		}
		return m.buffer[s : m.cursor+len(m.BlockCommentEnd)-1]
	}
	return nil
}

func (m *Marshaler) unmarshalInlineComment() []byte {
	if m.isLineCommentStart() {
		s := m.cursor
		for m.cursor < m.len {
			if m.isLineCommentEnd() {
				m.Inc(len(m.LineCommentEnd))
				break
			}
			m.Inc()
		}
		return m.buffer[s : m.cursor+len(m.LineCommentEnd)-1]
	}
	return nil
}

func (m *Marshaler) unmarshalNonData() []byte {
	data, s := false, m.cursor
	for m.cursor < m.len && !data {
		switch {
		case m.isBlockCommentStart():
			m.unmarshalCommentBlock()
		case m.isLineCommentStart():
			m.unmarshalInlineComment()
		case m.isSpace():
			m.unmarshalSpace()
		default:
			data = true
		}
	}
	return m.buffer[s:m.cursor]
}

func (m *Marshaler) UnmarshalTo(stop []byte) []byte {
	s := m.cursor
	e := stop[0]
	for m.cursor < m.len {
		if m.ByteIs(e) {
			if MatchBytes(m.buffer[m.cursor:m.cursor+len(stop)], stop) {
				m.Inc(len(stop))
				break
			}
		}
	}
	return m.buffer[s:m.cursor]
}

func (m *Marshaler) unmarshalSliceStart(ancestry []ancestor) (delim, end []byte, is bool) {
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

func (m *Marshaler) unmarshalMapStart(ancestry []ancestor) (delim, end []byte, is bool) {
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

func (m *Marshaler) Inc(i ...int) {
	if i == nil {
		m.cursor++
		return
	}
	m.cursor += i[0]
}

func (m *Marshaler) Byte() byte {
	if m.len == 0 {
		return 0
	}
	return m.buffer[m.cursor]
}

func (m *Marshaler) ByteIs(b byte) bool {
	return m.Byte() == b
}

func (m *Marshaler) isSpace() bool {
	return InBytes(m.buffer[m.cursor], m.Space)
}

func (m *Marshaler) isQuote() bool {
	return InBytes(m.buffer[m.cursor], m.Quote)
}

func (m *Marshaler) isEscape() bool {
	return InBytes(m.buffer[m.cursor], m.Escape)
}

func (m *Marshaler) isNull() bool {
	return m.isMatch(m.Null)
}

func (m *Marshaler) isKey() bool {
	i := m.cursor
	l := len(m.KeyEnd)
	for i < m.len {
		if m.buffer[i] == m.LineBreak[0] {
			return false
		}
		if MatchBytes(m.buffer[i:i+l], m.KeyEnd) {
			return true
		}
		i++
	}
	return false
}

func (m *Marshaler) isKeyEnd() bool {
	return InBytes(m.buffer[m.cursor], m.KeyEnd)
}

func (m *Marshaler) isBlockCommentStart() bool {
	return m.isMatch(m.BlockCommentStart)
}

func (m *Marshaler) isBlockCommentEnd() bool {
	return m.isMatch(m.BlockCommentEnd)
}
func (m *Marshaler) isLineCommentStart() bool {
	return m.isMatch(m.LineCommentStart)
}

func (m *Marshaler) isLineCommentEnd() bool {
	return m.isMatch(m.LineCommentEnd)
}

func (m *Marshaler) isMatch(b ...[]byte) bool {
	for _, s := range b {
		if m.len-m.cursor < len(s) {
			continue
		}
		if MatchBytes(m.buffer[m.cursor:m.cursor+len(s)], s) {
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

func EscapeBytes(b []byte, escape byte, chars ...byte) []byte {
	var escaped []byte
	for _, c := range b {
		if InBytes(c, chars) {
			escaped = append(escaped, escape)
		}
		escaped = append(escaped, c)
	}
	return escaped
}
