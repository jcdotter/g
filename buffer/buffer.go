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

package buffer

import (
	"errors"
	"strconv"
	"sync"
	"unsafe"
)

const defaultBufferSize = 1024

var (
	ErrEOF    = errors.New("EOF")
	ErrRange  = errors.New("index out of range")
	ErrLength = errors.New("invalid length")
)

// --------------------------------------------------------------------------- /
// Buffer Pool
// --------------------------------------------------------------------------- /

// Pool is a type-safe wrapper around a sync.Pool
// that provides a pool of Buffers.
var Pool = NewPool()

// BufferPool is a type-safe wrapper around a sync.Pool.
type bufferPool struct {
	p *sync.Pool
}

// NewPool constructs a new BufferPool.
func NewPool() *bufferPool {
	return &bufferPool{
		p: &sync.Pool{
			New: func() any { return New() },
		},
	}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p *bufferPool) Get() *Buffer {
	b := p.p.Get().(*Buffer)
	b.p = p
	return b
}

// Put returns a Buffer to the pool.
func (p *bufferPool) Put(b *Buffer) {
	b.Reset()
	p.p.Put(b)
}

// --------------------------------------------------------------------------- /
// Buffer
// --------------------------------------------------------------------------- /

// Buffer is a type-safe wrapper around a byte slice.
type Buffer struct {
	b []byte
	p *bufferPool
}

// New constructs a new Buffer.
func New() *Buffer {
	return &Buffer{b: make([]byte, 0, defaultBufferSize)}
}

// Reset resets the Buffer's byte slice.
func (b *Buffer) Reset() *Buffer {
	b.b = b.b[:0]
	return b
}

// Free returns the Buffer to the pool.
func (b *Buffer) Free() {
	b.p.Put(b)
}

// Grow doubles the Buffer's byte slice capacity.
func (b *Buffer) Grow() *Buffer {
	n := make([]byte, 0, cap(b.b)*2)
	copy(n, b.b)
	b.b = n
	return b
}

// --------------------------------------------------------------------------- /
// Write methods
// --------------------------------------------------------------------------- /

// Set sets the Buffer's byte slice.
func (b *Buffer) Set(bytes []byte) (int, error) {
	b.b = bytes
	return len(bytes), nil
}

// Write writes the given bytes to the Buffer's byte slice.
func (b *Buffer) Write(bytes []byte) (int, error) {
	b.b = append(b.b, bytes...)
	return len(bytes), nil
}

// WriteByte writes the given byte to the Buffer's byte slice.
func (b *Buffer) WriteByte(n byte) error {
	b.b = append(b.b, n)
	return nil
}

// WriteRune writes the given rune char to the Buffer's byte slice.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	bytes, n := RuneToBytes(r)
	b.b = append(b.b, bytes...)
	return n, nil
}

// WriteBytes writes the given bytes to the Buffer's byte slice.
func (b *Buffer) WriteBytes(bytes []byte) (int, error) {
	return b.Write(bytes)
}

// WriteByteSlices writes the given bytes to the Buffer's byte slice.
func (b *Buffer) WriteByteSlices(bytes ...[]byte) (int, error) {
	var l int
	for _, s := range bytes {
		b.b = append(b.b, s...)
		l += len(s)
	}
	return l, nil
}

// WriteBool writes the given bool to the Buffer's byte slice.
func (b *Buffer) WriteBool(n bool) (int, error) {
	s := strconv.FormatBool(n)
	b.b = append(b.b, s...)
	return len(s), nil
}

// WriteInt writes the given int to the Buffer's byte slice.
func (b *Buffer) WriteInt(n int) (int, error) {
	s := strconv.Itoa(n)
	b.b = append(b.b, s...)
	return len(s), nil
}

// WriteUint writes the given uint to the Buffer's byte slice.
func (b *Buffer) WriteUint(n uint) (int, error) {
	s := strconv.FormatUint(uint64(n), 10)
	b.b = append(b.b, s...)
	return len(s), nil
}

// WriteFloat writes the given float to the Buffer's byte slice.
func (b *Buffer) WriteFloat(n float64) (int, error) {
	s := strconv.FormatFloat(n, 'f', -1, 64)
	b.b = append(b.b, s...)
	return len(s), nil
}

// WriteString writes the given string to the Buffer's byte slice.
func (b *Buffer) WriteString(s string) (int, error) {
	b.b = append(b.b, s...)
	return len(s), nil
}

// WriteStrings writes the given strings to the Buffer's byte slice.
func (b *Buffer) WriteStrings(strings ...string) (int, error) {
	var l int
	for _, s := range strings {
		b.b = append(b.b, s...)
		l += len(s)
	}
	return l, nil
}

// Prepend prepends the given bytes to the Buffer's byte slice.
func (b *Buffer) Prepend(bytes []byte) (int, error) {
	b.b = append(bytes, b.b...)
	return len(bytes), nil
}

// PrependRune prepends the given rune to the Buffer's byte slice.
func (b *Buffer) PrependRune(r rune) (int, error) {
	bytes, n := RuneToBytes(r)
	b.b = append(bytes, b.b...)
	return n, nil
}

// PrependString prepends the given string to the Buffer's byte slice.
func (b *Buffer) PrependString(s string) (int, error) {
	bytes := []byte(s)
	b.b = append(bytes, b.b...)
	return len(bytes), nil
}

// Insert inserts the given bytes at the given index in the Buffer's byte slice.
func (b *Buffer) Insert(index int, bytes []byte) (int, error) {
	switch index = b.GetIndex(index); index {
	case 0:
		return b.Prepend(bytes)
	case len(b.b):
		return b.Write(bytes)
	default:
		b.b = append(b.b[:index], append(bytes, b.b[index:]...)...)
		return len(bytes), nil
	}
}

// InsertRune inserts the given rune at the given index in the Buffer's byte slice.
func (b *Buffer) InsertRune(index int, r rune) (int, error) {
	switch index, _ = RuneIndex(b.b, index); index {
	case 0:
		return b.PrependRune(r)
	case len(b.b):
		return b.WriteRune(r)
	default:
		bytes, n := RuneToBytes(r)
		b.b = append(b.b[:index], append(bytes, b.b[index:]...)...)
		return n, nil
	}
}

// InsertByte inserts the given byte at the given index in the Buffer's byte slice.
func (b *Buffer) InsertByte(index int, n byte) (int, error) {
	return b.Insert(index, []byte{n})
}

// InsertString inserts the given string at the given index in the Buffer's byte slice.
func (b *Buffer) InsertString(index int, s string) (int, error) {
	return b.Insert(index, []byte(s))
}

// Truncate truncates the Buffer's byte slice to the given index.
// If the given length is negative, the Buffer's byte slice is truncated
// to the given length from the end of the Buffer's byte slice.
// Byte slice is evaluated as []rune to allow for rune values.
func (b *Buffer) Truncate(length int) (int, error) {
	length, _ = RuneIndex(b.b, length)
	b.b = b.b[:length]
	return length, nil
}

// Delete removes the given length at the given index in the Buffer's byte slice.
// Byte slice is evaluated as []rune to allow for rune values.
func (b *Buffer) Delete(index, length int) (n int, err error) {
	if b.Len() == 0 {
		return 0, ErrEOF
	}
	if index, n, err = RangeRunes(b.b, index, length); err != nil {
		return
	}
	b.b = append(b.b[:index], b.b[index+n:]...)
	return
}

// Backspace removes the byte before the given index in the Buffer's byte slice.
// Byte slice is evaluated as []rune to allow for rune values.
func (b *Buffer) Backspace(index, length int) (int, error) {
	return b.Delete(index, -length)
}

// --------------------------------------------------------------------------- /
// Must write methods
// --------------------------------------------------------------------------- /

// MustSet sets the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustSet(bytes []byte) *Buffer {
	Must(b.Set(bytes))
	return b
}

// MustWrite writes the given bytes to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWrite(bytes []byte) *Buffer {
	Must(b.Write(bytes))
	return b
}

// MustWriteByte writes the given byte to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteByte(n byte) *Buffer {
	Must(0, b.WriteByte(n))
	return b
}

// MustWriteBytes writes the given bytes to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteBytes(bytes []byte) *Buffer {
	Must(b.Write(bytes))
	return b
}

// MustWriteBool writes the given bool to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteBool(n bool) *Buffer {
	Must(b.WriteBool(n))
	return b
}

// MustWriteInt writes the given int to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteInt(n int) *Buffer {
	Must(b.WriteInt(n))
	return b
}

// MustWriteUint writes the given uint to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteUint(n uint) *Buffer {
	Must(b.WriteUint(n))
	return b
}

// MustWriteFloat writes the given float to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteFloat(n float64) *Buffer {
	Must(b.WriteFloat(n))
	return b
}

// MustWriteString writes the given string to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteString(s string) *Buffer {
	Must(b.WriteString(s))
	return b
}

// MustWriteStrings writes the given strings to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustWriteStrings(strings ...string) *Buffer {
	Must(b.WriteStrings(strings...))
	return b
}

// MustPrepend prepends the given bytes to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustPrepend(bytes []byte) *Buffer {
	Must(b.Prepend(bytes))
	return b
}

// MustPrependString prepends the given string to the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustPrependString(s string) *Buffer {
	Must(b.PrependString(s))
	return b
}

// MustInsert inserts the given bytes at the given index in the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustInsert(index int, bytes []byte) *Buffer {
	Must(b.Insert(index, bytes))
	return b
}

// MustInsertByte inserts the given byte at the given index in the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustInsertByte(index int, n byte) *Buffer {
	Must(b.InsertByte(index, n))
	return b
}

// MustInsertString inserts the given string at the given index in the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustInsertString(index int, s string) *Buffer {
	Must(b.InsertString(index, s))
	return b
}

// MustTruncate truncates the Buffer's byte slice to the given length
// and panics if an error occurs.
func (b *Buffer) MustTruncate(length int) *Buffer {
	Must(b.Truncate(length))
	return b
}

// MustDelete removes the given bytes from the Buffer's byte slice at the given index
// and panics if an error occurs.
func (b *Buffer) MustDelete(index, length int) *Buffer {
	Must(b.Delete(index, length))
	return b
}

// MustBackspace removes the byte before the given index in the Buffer's byte slice
// and panics if an error occurs.
func (b *Buffer) MustBackspace(index, length int) *Buffer {
	Must(b.Backspace(index, length))
	return b
}

// --------------------------------------------------------------------------- /
// Read methods
// --------------------------------------------------------------------------- /

// Read reads from the Buffer's byte slice into the given bytes.
func (b *Buffer) Read(bytes []byte) (int, error) {
	if len(b.b) == 0 {
		return 0, ErrEOF
	}
	n := copy(bytes, b.b)
	b.b = b.b[n:]
	return n, nil
}

// --------------------------------------------------------------------------- /
// Must read methods
// --------------------------------------------------------------------------- /

// MustRead reads from the Buffer's byte slice into the given bytes
// and panics if an error occurs.
func (b *Buffer) MustRead(bytes []byte) *Buffer {
	Must(b.Read(bytes))
	return b
}

// Buffer returns the Buffer's byte slice.
func (b *Buffer) Buffer() []byte {
	return b.b
}

// Copy returns a copy of the Buffer.
func (b *Buffer) Copy() *Buffer {
	if b == nil {
		return nil
	}
	n := b.p.Get()
	n.b = append(n.b, b.b...)
	return n
}

// Bytes returns a copy Buffer's byte slice.
func (b *Buffer) Bytes() []byte {
	bytes := make([]byte, len(b.b))
	copy(bytes, b.b)
	return bytes
}

// String returns the Buffer's byte slice as a string.
func (b *Buffer) String() string {
	return string(b.b)
}

// Len returns the length of the Buffer's byte slice.
func (b *Buffer) Len() int {
	return len(b.b)
}

// Get returns the byte at the given index in the Buffer's byte slice.
func (b *Buffer) Get(index int) byte {
	return b.b[index]
}

// GetIndex returns the index of byte i in the Buffer's byte slice. If i exceeds the Buffer's length,
// the Buffer's length is returned. If i is negative, the Buffer's length minus i is returned.
func (b *Buffer) GetIndex(i int) (index int) {
	switch {
	case i > len(b.b):
		return len(b.b)
	case -i >= len(b.b):
		return 0
	case i < 0:
		return len(b.b) + i
	default:
		return i
	}
}

// GetRune returns the number of bytes in the Buffer's byte slice
// required to represent the rune at the given index.
func (b *Buffer) GetRune(index int) (r rune, n int, err error) {
	i := index
	for ; i < len(b.b); i++ {
		if b.b[i] < 128 {
			break
		}
	}
	bytes := make([]byte, 4)
	copy(bytes, b.b[index:i])
	r = *(*rune)(unsafe.Pointer(&bytes[0]))
	return r, i - index + 1, nil
}

// Runes returns a slice of runes from the Buffer's byte slice.
func (b *Buffer) Runes() []rune {
	return []rune(b.String())
}

// --------------------------------------------------------------------------- /
// Rune methods
// --------------------------------------------------------------------------- /

// RuneIndex returns the byte index and byte len of rune i in the Buffer's byte slice.
// If i exceeds the Buffer's length, the Buffer's length is returned. If i is negative,
// the Buffer's length minus i is returned.
func RuneIndex(bytes []byte, i int) (byteIndex int, byteLength int) {
	r := []rune(string(bytes))
	l := len(r)
	switch {
	case i >= l:
		return len(bytes), 0
	case -i >= l:
		i = 0
	case i < 0:
		i = l + i
	}
	if i == 0 {
		return 0, RuneLen(r[0])
	}
	return RunesLen(r[:i]), RuneLen(r[i])
}

// RuneLen returns the number of bytes required to represent the given rune.
func RuneLen(r rune) int {
	switch {
	case r < 128:
		return 1
	case r < 2048:
		return 2
	case r < 65536:
		return 3
	default:
		return 4
	}
}

// RunesLen returns the number of bytes required to represent the given runes.
func RunesLen(runes []rune) (n int) {
	for _, r := range runes {
		n += RuneLen(r)
	}
	return
}

// RuneToBytes converts a rune to a byte slice.
func RuneToBytes(r rune) (b []byte, n int) {
	switch n = RuneLen(r); n {
	case 1:
		b = []byte{byte(r)}
	case 2:
		b = make([]byte, n)
		*(*uint16)(unsafe.Pointer(&b[0])) = uint16(r)
	case 3:
		b = make([]byte, n)
		*(*uint32)(unsafe.Pointer(&b[0])) = uint32(r)
	default:
		b = make([]byte, n)
		*(*rune)(unsafe.Pointer(&b[0])) = r
	}
	return
}

// RangeRunes evaluates and returns a valid index and length of the given []byte,
// adjusted for rune values and allowing for negative index and length values.
func RangeRunes(bytes []byte, runeIndex, runesLen int) (byteIndex int, bytesLen int, err error) {
	if runesLen == 0 {
		return 0, 0, ErrLength
	}
	r := []rune(string(bytes))
	runeIndex, runesLen, err = Range(len(r), runeIndex, runesLen)
	return RunesLen(r[:runeIndex]), RunesLen(r[runeIndex : runeIndex+runesLen]), err
}

// --------------------------------------------------------------------------- /
// Helps
// --------------------------------------------------------------------------- /

// Must panics if the given error is not nil.
func Must(_ int, err error) {
	if err != nil {
		panic(err)
	}
}

// Range evaluates and returns a valid index and length of the given []byte,
// allowing for negative index and length values.
func Range(bytes, index, length int) (int, int, error) {
	if length == 0 {
		return 0, 0, ErrLength
	}
	switch {
	case index >= 0:
		switch {
		case length > 0:
			switch {
			case index >= bytes:
				return 0, 0, ErrRange
			case index+length > bytes:
				length = bytes - index
			}
		case index == 0:
			return Range(bytes, bytes+length, -length)
		case index > 0:
			return Range(bytes, index+length, -length)
		}
	case index < 0:
		switch {
		case length > 0:
			index = bytes + index
			switch {
			case index >= 0:
				return Range(bytes, index, length)
			case -index > length:
				return 0, 0, ErrRange
			default:
				return Range(bytes, 0, length+index)
			}
		case length < 0:
			return Range(bytes, index+length, -length)
		}
	}
	return index, length, nil
}
