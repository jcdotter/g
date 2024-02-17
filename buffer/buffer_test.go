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
	"testing"

	"github.com/jcdotter/go/buffer/test"
)

var config = &test.Config{
	//PrintTest:   true,
	PrintFail:   true,
	PrintTrace:  true,
	PrintDetail: true,
	FailFatal:   true,
	Msg:         "%s",
}

func TestAll(t *testing.T) {
	gt := test.New(t, config)

	var b = New()
	gt.NotEqual(nil, b, "buffer should not be nil")
	gt.NotEqual(nil, b.b, "buffer slice should not be nil")
	gt.NotEqual(nil, b.p, "buffer pool should not be nil")
	gt.Equal(0, b.Len(), "buffer should have a length of 0")
	gt.Equal(defaultBufferSize, cap(b.b), "buffer should have a capacity of 1024")

	b.Reset().MustWriteByte('t')
	gt.Equal([]byte{'t'}, b.Bytes(), "buffered byte should be equal to 't'")

	v := []byte("1234567890")
	b.Reset().MustWriteBytes(v)
	gt.Equal(v, b.Bytes(), "buffered bytes should be equal to '1234567890'")

	b.Reset().MustWriteBool(true)
	gt.Equal([]byte("true"), b.Bytes(), "buffered bool should be equal to 'true'")

	b.Reset().MustWriteInt(1234567890)
	gt.Equal(v, b.Bytes(), "buffered int should be equal to '1234567890'")

	b.Reset().MustWriteUint(1234567890)
	gt.Equal(v, b.Bytes(), "buffered uint should be equal to '1234567890'")

	b.Reset().MustWriteFloat(1234567890)
	gt.Equal(v, b.Bytes(), "buffered float should be equal to '1234567890'")

	b.Reset().MustWriteString("1234567890")
	gt.Equal(v, b.Bytes(), "buffer string should be equal to '1234567890'")

	b.Reset().MustWriteStrings("12345", "67890")
	gt.Equal(v, b.Bytes(), "buffer string should be equal to '1234567890'")

	b.Reset().MustWriteString("1234567890")
	gt.Equal(10, b.Len(), "buffer length should be equal to 10")
	gt.Equal(1024, cap(b.b), "buffer capacity should be equal to 1024")
	gt.Equal(9, b.GetIndex(-1), "buffer index -1 index should be equal to 9")

	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("123456789"), b.MustDelete(0, 1).Bytes(), "buffer.Delete(0,1) should be equal to '123456789'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("012345678"), b.MustDelete(0, -1).Bytes(), "buffer.Delete(0,-1) should be equal to '012345678'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("01234569"), b.MustDelete(-1, -2).Bytes(), "buffer.Delete(-1,-2) should be equal to '01234569'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("01456789"), b.MustDelete(2, 2).Bytes(), "buffer.Delete(2,2) should be equal to '123456789'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("23456789"), b.MustDelete(-15, 7).Bytes(), "buffer.Delete(-15,7) should be equal to '23456789'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("0123"), b.MustTruncate(4).Bytes(), "buffer.Trucate(4) should be equal to '0123'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("012345"), b.MustTruncate(-4).Bytes(), "buffer.Trucate(4) should be equal to '012345'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("012349"), b.MustBackspace(-1, 4).Bytes(), "buffer.Backspace(-1,4) should be equal to '012349'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("01234567"), b.MustBackspace(0, 2).Bytes(), "buffer.Backspace(0,2) should be equal to '01234567'")
	b.Reset().MustWriteString("0123456789")
	gt.Equal([]byte("01234567"), b.MustBackspace(0, 2).Bytes(), "buffer.Backspace(0,2) should be equal to '01234567'")

	b.Reset().MustWriteString("enseñé 123")
	gt.Equal(12, b.Len(), "buffer len should be equal to 12")
	gt.Equal([]byte("enseñ"), b.MustTruncate(-5).Bytes(), "buffer should be equal to 'enseñ'")
}
