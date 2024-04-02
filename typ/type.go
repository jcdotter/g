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

package typ

import (
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// ------------------------------------------------------------ /
// Type IMPLEMENTATION
// custom implementation of golang source code: reflect.Type
// with expanded functionality

var (
	__ttime       = (any)(time.Time{})
	__timefields  = (*structType)(unsafe.Pointer((*Iface)(unsafe.Pointer(&__ttime)).Type)).fields
	__timefield0b = __timefields[0].name.bytes
	__timefield0t = __timefields[0].typ
	__timefield1b = __timefields[1].name.bytes
	__timefield1t = __timefields[1].typ
	__timefield2b = __timefields[2].name.bytes
	__timefield2t = __timefields[2].typ
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
)

type Type struct {
	size       uintptr
	ptrdata    uintptr // number of bytes in the type that can contain pointers
	hash       uint32  // hash of type; avoids computation in hash tables
	tflag      tflag   // extra type information flags
	align      uint8   // alignment of variable with this type
	fieldAlign uint8   // alignment of struct field with this type
	kind       uint8   // enumeration for C
	equal      func(unsafe.Pointer, unsafe.Pointer) bool
	gcdata     *byte   // garbage collection data
	str        nameOff // string form
	ptrToThis  typeOff // type for pointer to this type, may be zero
}

type nameOff int32 // offset to a name
type typeOff int32 // offset to a *Type

// TypeOf returns the typ of value a
func TypeOf(a any) *Type {
	return *(**Type)(unsafe.Pointer(&a))
}

// FromReflectType returns the typ of reflect.Type t
func FromReflectType(t reflect.Type) *Type {
	a := (any)(t)
	return (*Type)((*Iface)(unsafe.Pointer(&a)).Pointer)
}

// New returns an empty pointer to a new value of the Type
func (t *Type) New() Value {
	if t != nil {
		return Iface{t.PtrType(), unsafe_New(t), flag(POINTER)}.Value()
	}
	panic("call to New on nil type")
}

// New returns a pointer to a new (non nil) value of the Type
func (t *Type) NewValue() Value {
	n := t.New()
	switch t.Kind() {
	case MAP:
		*(*unsafe.Pointer)(n.Pointer()) = makemap(t, 0, nil)
	case POINTER:
		*(*unsafe.Pointer)(n.Pointer()) = t.Elem().NewValue().Pointer()
	case SLICE:
		t := (*sliceType)(unsafe.Pointer(t)).elem
		p := n.Pointer()
		*(*unsafe.Pointer)(&p) = unsafe.Pointer(&sliceHeader{unsafe_NewArray(t, 0), 0, 0})
	}
	return n
}

// Reflect returns the reflect.Type of the Type
func (t *Type) Reflect() reflect.Type {
	return toType(t)
}

// PtrType returns a new Type of a pointer to the Type
func (t *Type) PtrType() *Type {
	return FromReflectType(reflect.PtrTo(toType(t)))
}

// IfaceIndir returns true if the Type is an indirect value
func (t *Type) IfaceIndir() bool {
	return t.kind&KindDirectIface == 0
}

// flag returns the flag of the Type
func (t *Type) flag() flag {
	f := flag(t.kind & kindMask)
	if t.IfaceIndir() {
		f |= flagIndir
	}
	return f
}

// RefKind returns the kind of the Type synonomous with reflect.Kind
func (t *Type) Kind() byte {
	return t.kind & kindMask
}

// KindX returns the typ kind of the Type
// to include aliases and special types
func (t *Type) KindX() byte {
	k := t.kind & kindMask
	switch k {
	case SLICE: // check if byte array
		ek := (*sliceType)(unsafe.Pointer(t)).elem.kind & kindMask
		if ek == 8 || ek == 10 { // []byte or []rune
			return BINARY
		}
	case ARRAY: // check if uuid
		a := (*arrayType)(unsafe.Pointer(t))
		if a.len == 16 && a.elem.kind&kindMask == 8 { // [16]byte
			return UUID
		}
	case STRUCT: // check if time
		fs := (*structType)(unsafe.Pointer(t)).fields
		if len(fs) == 3 {
			if fs[0].name.bytes != __timefield0b || fs[0].typ != __timefield0t ||
				fs[1].name.bytes != __timefield1b || fs[1].typ != __timefield1t ||
				fs[2].name.bytes != __timefield2b || fs[2].typ != __timefield2t {
				return STRUCT
			} else {
				return TIME
			}
		}
	}
	return k
}

// Elem returns the Type of the element of the Type
func (t *Type) Elem() *Type {
	switch t.Kind() {
	case ARRAY:
		return (*arrayType)(unsafe.Pointer(t)).elem
	case MAP:
		return (*mapType)(unsafe.Pointer(t)).elem
	case POINTER:
		return (*ptrType)(unsafe.Pointer(t)).elem
	case SLICE:
		return (*sliceType)(unsafe.Pointer(t)).elem
	}
	return t
}

func (t *Type) DeepPtrElem() *Type {
	for t.Kind() == POINTER {
		t = (*ptrType)(unsafe.Pointer(t)).elem
	}
	return t
}

// IsData returns true if the Type stores data
// which includes Array, Chan, Map, Slice, Struct, Bytes, Interface
// or is a pointer to one these types
func (t *Type) IsData() bool {
	k := t.Kind()
	return k == ARRAY || k == CHAN || k == MAP || k == SLICE || k == STRUCT || k == BINARY || k == INTERFACE ||
		(k == POINTER && t.DeepPtrElem().IsData())
}

func (t *Type) HasDataElem() bool {
	switch t.Kind() {
	case POINTER:
		return t.DeepPtrElem().HasDataElem()
	case STRUCT:
		return t.HasDataField()
	case INTERFACE:
		return true
	default:
		return t.Elem().IsData()
	}
}

func (t *Type) IsError() bool {
	if t.Kind() == INTERFACE {
		return t.Reflect().Implements(errorType)
	}
	return false
}

// String returns the string representation of the Type
func (t *Type) String() string {
	return t.Name()
}

// Name returns the name of the Type
func (t *Type) Name() string {
	n := name{(*byte)(resolveNameOff(unsafe.Pointer(t), int32(t.str)))}.name()
	if t.Kind() != POINTER {
		n = n[1:]
	}
	return n
}

// NameShort returns the short name of the Type
// excluding the package path, module name and pointer indicator
func (t *Type) NameShort() string {
	n := t.Name()
	return string(n[strings.LastIndex(t.Name(), ".")+1:])
}

// SoftMatch evaluates whether typ matches the data type structure of Type t
// although maybe not identical
func (t *Type) SoftMatch(typ *Type, ancestry ...*Type) bool {
	if typ.InTypes(ancestry...) {
		return true
	}
	ancestry = append(ancestry, t)
	if k := t.Kind(); k == typ.Kind() {
		switch k {
		default:
			return true
		case POINTER, ARRAY, SLICE:
			return t.Elem().SoftMatch(typ.Elem(), ancestry...)
		case STRUCT:
			return StructTypeMatch(t, typ, ancestry...)
		case MAP:
			tt := (*mapType)(unsafe.Pointer(t))
			yy := (*mapType)(unsafe.Pointer(typ))
			return tt.key.SoftMatch(yy.key, ancestry...) && tt.elem.SoftMatch(yy.elem, ancestry...)
		}
	}
	return false
}

func (t *Type) InTypes(types ...*Type) bool {
	for _, s := range types {
		if t == s {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------ /
// STURCTURED TypeS
// implementation of golang types for data structures:
// array, map, ptr, slice, string, struct, field, interface
// ------------------------------------------------------------ /

type arrayType struct {
	Type
	elem  *Type // array element type
	slice *Type // slice type
	len   uintptr
}

type funcType struct {
	Type
	inCount  uint16
	outCount uint16
}

type mapType struct {
	Type
	key    *Type // map key type
	elem   *Type // map element (value) type
	bucket *Type // internal bucket structure
	// function for hashing keys (ptr to key, seed) -> hash
	hasher     func(unsafe.Pointer, uintptr) uintptr
	keysize    uint8  // size of key slot
	valuesize  uint8  // size of value slot
	bucketsize uint16 // size of bucket
	flags      uint32
}

type bmap struct {
	_ [bucketCnt]uint8
}

const (
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits
	dataOffset    = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)
)

type hiter struct {
	_ unsafe.Pointer    // key
	_ unsafe.Pointer    // elem
	_ unsafe.Pointer    // t
	_ unsafe.Pointer    // h
	_ unsafe.Pointer    // buckets
	_ unsafe.Pointer    // bptr
	_ *[]unsafe.Pointer // overflow
	_ *[]unsafe.Pointer // oldoverflow
	_ uintptr           // startBucket
	_ uint8             // offset
	_ bool              // wrapped
	_ uint8             // B
	_ uint8             // i
	_ uintptr           // bucket
	_ uintptr           // checkBucket
}

type ptrType struct {
	Type
	elem *Type
}

type sliceType struct {
	Type
	elem *Type // slice element type
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

type stringHeader struct {
	Data uintptr
	Len  int
}

type structType struct {
	Type
	pkgPath name
	fields  []FieldType // sorted by offset
}

type FieldType struct {
	name   name    // name is always non-empty
	typ    *Type   // type of field
	offset uintptr // byte offset of field
}

type interfaceType struct {
	Type
	PkgPath name      // import path
	Methods []Imethod // sorted by hash
}

func (i *interfaceType) NumMethod() int {
	return len(i.Methods)
}

type Imethod struct {
	_ nameOff
	_ typeOff
}

// ------------------------------------------------------------ /
// STRUCT Type IMPLEMENTATION
// custom implementation of golang struct type
// ------------------------------------------------------------ /

// IsStruct returns true if the Type is a struct
func (t *Type) IsStruct() bool {
	return t.Kind() == STRUCT
}

// PkgPath returns the package path of a struct Type
func (t *Type) PkgPath() string {
	return (*structType)(unsafe.Pointer(t)).pkgPath.name()
}

// NumField returns the number of fields in a struct Type
func (t *Type) NumField() int {
	return len((*structType)(unsafe.Pointer(t)).fields)
}

// Field returns the Type of the field at index i in a struct Type
func (t *Type) Field(i int) *FieldType {
	// fields array pointer = structType pointer + size of Type (48) + size of name (8)
	// field pointer = fields array pointer + size of field (24) * index
	return (*FieldType)(offseti(*(*unsafe.Pointer)(offset(unsafe.Pointer(t), 56)), i*24))
}

// FieldByName returns the Type of the field with name in a struct Type
func (t *Type) FieldByName(name string) *FieldType {
	fs := (*structType)(unsafe.Pointer(t)).fields
	for i, f := range fs {
		if f.name.name() == name {
			return t.Field(i)
		}
	}
	return nil
}

// FieldByTag returns the Type of the field with tag value in a struct Type
func (t *Type) FieldByTag(tag string, value string) *FieldType {
	fs := (*structType)(unsafe.Pointer(t)).fields
	for i, f := range fs {
		v := f.name.tagValue(tag)
		if v == value {
			return t.Field(i)
		}
	}
	return nil
}

// FieldByIndex returns the Type of the field at index in a struct Type
func (t *Type) FieldByIndex(index []int) *FieldType {
	switch len(index) {
	case 0:
		return nil
	case 1:
		return t.Field(index[0])
	default:
		return t.Field(index[0]).typ.FieldByIndex(index[1:])
	}
}

// FieldName returns the name of the field at index i in a struct Type
func (t *Type) FieldName(i int) string {
	return (*structType)(unsafe.Pointer(t)).fields[i].name.name()
}

// FieldIndex returns the index of the field with name in a struct Type
func (t *Type) FieldIndex(name string) int {
	fs := (*structType)(unsafe.Pointer(t)).fields
	for i, f := range fs {
		if f.name.name() == name {
			return i
		}
	}
	return 0
}

// IndexTag returns the tag of the field at index i in a struct Type
func (t *Type) IndexTag(i int) string {
	return (*structType)(unsafe.Pointer(t)).fields[i].name.tag()
}

// FieldTag returns the tag of the field with name in a struct Type
func (t *Type) FieldTag(name string) string {
	fs := (*structType)(unsafe.Pointer(t)).fields
	for _, f := range fs {
		if f.name.name() == name {
			return f.name.tag()
		}
	}
	return ""
}

// IndexTagValue returns the value of the tag of the field at index i in a struct Type
func (t *Type) IndexTagValue(i int, tag string) string {
	return t.Field(i).name.tagValue(tag)
}

// FieldTagValue returns the value of the tag of the field with name in a struct Type
func (t *Type) FieldTagValue(name string, tag string) string {
	return t.FieldByName(name).name.tagValue(tag)
}

// TagValues returns a slice of string values for tag across fields in a struct Type
func (t *Type) TagValues(tag string) (vals []string, has bool) {
	fs := (*structType)(unsafe.Pointer(t)).fields
	vals = make([]string, len(fs))
	has = true
	for i, f := range fs {
		vals[i] = f.name.tagValue(tag)
		if vals[i] == "" {
			has = false
			break
		}
	}
	return
}

// ForFields iterates over the fields of a struct Type and calls
// the function f with the index and Type of each field
func (t *Type) ForFields(f func(i int, f *FieldType) (brake bool)) {
	for i := range (*structType)(unsafe.Pointer(t)).fields {
		if brake := f(i, t.Field(i)); brake {
			break
		}
	}
}

// HasDataField returns true if the struct Type has a field with a data type
// of array, chan, map, slice, struct, bytes or interface
func (t *Type) HasDataField() bool {
	has := false
	t.ForFields(func(i int, f *FieldType) (brake bool) {
		if f.typ.IsData() {
			has = true
			return true
		}
		return
	})
	return has
}

// StructTypeMatch compairs the structure of 2 structs
func StructTypeMatch(x, y *Type, ancestry ...*Type) bool {
	if x.IsStruct() && y.IsStruct() {
		xfs := (*structType)(unsafe.Pointer(x)).fields
		yfs := (*structType)(unsafe.Pointer(y)).fields
		if len(xfs) == len(yfs) {
			for i, xf := range xfs {
				yf := yfs[i]
				if xf.name.bytes != yf.name.bytes || !xf.typ.SoftMatch(yf.typ, ancestry...) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// ------------------------------------------------------------ /
// FIELD Type IMPLEMENTATION
// custom implementation of golang struct field type
// ------------------------------------------------------------ /

func (f *FieldType) Type() *Type {
	return f.typ
}

func (f *FieldType) String() string {
	return f.Name()
}

func (f *FieldType) Name() string {
	return f.name.name()
}

func (f *FieldType) Tag() string {
	return f.name.tag()
}

func (f *FieldType) Tags() map[string]string {
	return f.name.tags()
}

func (f *FieldType) TagValue(tag string) string {
	return f.name.tagValue(tag)
}

func (f *FieldType) Offset() uintptr {
	return f.offset
}

// ------------------------------------------------------------ /
// FUNC Type IMPLEMENTATION
// custom implementation of golang func type
// ------------------------------------------------------------ /

// IsFunc returns true if the Type is a func
func (t *Type) IsFunc() bool {
	return t.Kind() == FUNC
}

// NumIn returns the number of input parameters in a func Type
func (t *Type) NumIn() int {
	return int((*funcType)(unsafe.Pointer(t)).inCount)
}

// NumOut returns the number of output parameters in a func Type
func (t *Type) NumOut() int {
	return int((*funcType)(unsafe.Pointer(t)).outCount)
}

// In returns the Type of the input parameter at index i in a func Type
func (t *Type) In(i int) *Type {
	return (*funcType)(unsafe.Pointer(t)).in()[i]
}

// Out returns the Type of the output parameter at index i in a func Type
func (t *Type) Out(i int) *Type {
	return (*funcType)(unsafe.Pointer(t)).out()[i]
}

// in returns a slice of Types of the input parameters in a func Type
func (t *funcType) in() []*Type {
	uadd := unsafe.Sizeof(*t)
	if t.tflag&tflagUncommon != 0 {
		uadd += 32 //unsafe.Sizeof(uncommonType{})
	}
	if t.inCount == 0 {
		return nil
	}
	return (*[1 << 20]*Type)(add(unsafe.Pointer(t), uadd))[:t.inCount:t.inCount]
}

// out returns a slice of Types of the output parameters in a func Type
func (t *funcType) out() []*Type {
	uadd := unsafe.Sizeof(*t)
	if t.tflag&tflagUncommon != 0 {
		uadd += 32 //size of uncommonType
	}
	outCount := t.outCount & (1<<15 - 1)
	if outCount == 0 {
		return nil
	}
	return (*[1 << 20]*Type)(add(unsafe.Pointer(t), uadd))[t.inCount : t.inCount+outCount : t.inCount+outCount]
}

// ------------------------------------------------------------ /
// NAME IMPLEMENTATION
// custom implementation of golang source code: name
// with expanded functionality
// ------------------------------------------------------------ /

type name struct {
	bytes *byte
}

func (n name) data(off int) *byte {
	return (*byte)(add(unsafe.Pointer(n.bytes), uintptr(off)))
}

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func (n name) readVarint(off int) (int, int) {
	v := 0
	for i := 0; ; i++ {
		x := *n.data(off + i)
		v += int(x&0x7f) << (7 * i)
		if x&0x80 == 0 {
			return i + 1, v
		}
	}
}

func (n name) name() string {
	if n.bytes == nil {
		return ""
	}
	i, l := n.readVarint(1)
	return unsafe.String(n.data(1+i), l)
}

func (n name) hasTag() bool {
	return (*n.bytes)&(1<<1) != 0
}

func (n name) tag() string {
	if !n.hasTag() {
		return ""
	}
	i, l := n.readVarint(1)
	i2, l2 := n.readVarint(1 + i + l)
	return unsafe.String(n.data(1+i+l+i2), l2)
}

func (n name) tagValue(tag string) (value string) {
	t := n.tag()
	l := len(t)
	var name string
	var val bool
	for i := 0; i < l; i++ {
		if name, i, val = n.parseTagNameAt(tag, i); !val {
			if name == tag {
				return "true"
			}
			goto next
		}
		if value, i = n.parseTagValueAt(tag, i); name == tag {
			return value
		}
	next:
	}
	return ""
}

func (n name) tags() (tags map[string]string) {
	if !n.hasTag() {
		return
	}
	t := n.tag()
	l := len(t)
	tags = make(map[string]string)
	var name string
	var val bool
	for i := 0; i < l; i++ {
		if name, i, val = n.parseTagNameAt(t, i); !val {
			if name == t {
				tags[name] = "true"
			}
			goto next
		}
		tags[name], i = n.parseTagValueAt(t, i)
	next:
	}
	return

}

func (n name) parseTagNameAt(tag string, at int) (name string, end int, val bool) {
	for end = at; end < len(tag); end++ {
		switch tag[end] {
		case ':':
			val = true
			goto out
		case ',':
			goto out
		}
	}
out:
	name = tag[at:end]
	return
}

func (n name) parseTagValueAt(tag string, at int) (value string, end int) {
	if tag[at] == '"' {
		at++
		for end = at; end < len(tag); end++ {
			if tag[end] == '"' && tag[end-1] != '\\' {
				break
			}
		}
		value = tag[at:end]
		end++
		return
	}
	return "", at
}
