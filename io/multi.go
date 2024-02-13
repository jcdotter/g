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

package io

import (
	"errors"
	"io"
	"os"
	"reflect"
	"unsafe"
)

// io multi type for evaluating multiplexed readers and writers
var (
	multiReaderType     = reflect.TypeOf(io.MultiReader())
	multiWriterType     = reflect.TypeOf(io.MultiWriter())
	multiReadWriterType = reflect.TypeOf(&multiReadWriter{})
	ErrEOF              = errors.New("EOF")
	ErrShortWrite       = errors.New("short write")
)

// Reader mimics the io.Reader interface.
// See https://pkg.go.dev/io#Reader
type Reader interface {
	Read(p []byte) (n int, err error)
}

type multiReader struct {
	readers []io.Reader
}

func (r *multiReader) Read(p []byte) (n int, err error) {
	return 0, ErrEOF
}

func (r *multiReader) WriteTo(w Writer) (sum int64, err error) {
	return 0, ErrEOF
}

func ReaderFiles(r io.Reader) (fs []*os.File) {
	if f, is := r.(*os.File); is {
		fs = []*os.File{f}
	} else if reflect.TypeOf(r) == multiReaderType {
		mr := *(**multiReader)(unsafe.Pointer(uintptr(unsafe.Pointer(&r)) + 8))
		for _, x := range mr.readers {
			fs = append(fs, ReaderFiles(x)...)
		}
	}
	return
}

// Writer mimics the io.Writer interface.
// See https://pkg.go.dev/io#Writer
type Writer interface {
	Write(p []byte) (n int, err error)
}

type multiWriter struct {
	writers []io.Writer
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func WriterFiles(w io.Writer) (fs []*os.File) {
	if f, is := w.(*os.File); is {
		fs = []*os.File{f}
	} else if reflect.TypeOf(w) == multiWriterType {
		mw := *(**multiWriter)(unsafe.Pointer(uintptr(unsafe.Pointer(&w)) + 8))
		for _, x := range mw.writers {
			fs = append(fs, WriterFiles(x)...)
		}
	}
	return
}

// ReadWriter mimics the io.ReadWriter interface.
// See https://pkg.go.dev/io#ReadWriter
type ReadWriter interface {
	Reader
	Writer
}

// multiReadWriter mimics the io.ReadWriter interface.
// See https://pkg.go.dev/io#ReadWriter
type multiReadWriter struct {
	readWriters []io.ReadWriter
}

// See https://cs.opensource.google/go/go/+/refs/tags/go1.21rc4:src/io/multi.go
func (mr *multiReadWriter) Read(p []byte) (n int, err error) {
	for len(mr.readWriters) > 0 {
		if len(mr.readWriters) == 1 {
			if r, ok := mr.readWriters[0].(*multiReadWriter); ok {
				mr.readWriters = r.readWriters
				continue
			}
		}
		n, err = mr.readWriters[0].Read(p)
		if err == ErrEOF {
			mr.readWriters = mr.readWriters[1:]
		}
		if n > 0 || err != ErrEOF {
			if err == ErrEOF && len(mr.readWriters) > 0 {
				err = nil
			}
			return
		}
	}
	return 0, ErrEOF
}

// See https://cs.opensource.google/go/go/+/refs/tags/go1.21rc4:src/io/multi.go
func (mr *multiReadWriter) WriteTo(w Writer) (sum int64, err error) {
	return mr.writeToWithBuffer(w, make([]byte, 1024*32))
}

// See https://cs.opensource.google/go/go/+/refs/tags/go1.21rc4:src/io/multi.go
func (mr *multiReadWriter) writeToWithBuffer(w Writer, buf []byte) (sum int64, err error) {
	for i, r := range mr.readWriters {
		var n int64
		if subMr, ok := r.(*multiReadWriter); ok {
			n, err = subMr.writeToWithBuffer(w, buf)
		} else {
			n, err = copyBuffer(w, r, buf)
		}
		sum += n
		if err != nil {
			mr.readWriters = mr.readWriters[i:]
			return sum, err
		}
		mr.readWriters[i] = nil
	}
	mr.readWriters = nil
	return sum, nil
}

// See https://cs.opensource.google/go/go/+/refs/tags/go1.21rc4:src/io/multi.go
func (mw *multiReadWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.readWriters {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = ErrShortWrite
			return
		}
	}
	return len(p), nil
}

func (mw *multiReadWriter) WriteString(s string) (n int, err error) {
	return mw.Write([]byte(s))
}

// MultiReadWriter mimics the io.MultiWriter and io.MultiReader func.
// See https://pkg.go.dev/io#MultiWriter
func MultiReadWriter(readwriters ...io.ReadWriter) io.ReadWriter {
	return &multiReadWriter{readWriterSlice(readwriters...)}
}

func AppendReadWriter(mw io.ReadWriter, readwriters ...io.ReadWriter) io.ReadWriter {
	return &multiReadWriter{append(readWriterSlice(mw), readWriterSlice(readwriters...)...)}
}

func readWriterSlice(readwriters ...io.ReadWriter) []io.ReadWriter {
	mrw := make([]io.ReadWriter, len(readwriters))
	for _, rw := range readwriters {
		if smrw, ok := rw.(*multiReadWriter); ok {
			mrw = append(mrw, readWriterSlice(smrw.readWriters...)...)
		} else {
			mrw = append(mrw, rw)
		}
	}
	return mrw
}

func ReadWriterFiles(r io.ReadWriter) (fs []*os.File) {
	if f, is := r.(*os.File); is {
		fs = []*os.File{f}
	} else if reflect.TypeOf(r) == multiReadWriterType {
		mrw := *(**multiReadWriter)(unsafe.Pointer(uintptr(unsafe.Pointer(&r)) + 8))
		for _, x := range mrw.readWriters {
			fs = append(fs, ReadWriterFiles(x)...)
		}
	}
	return
}

//go:noescape
//go:linkname copyBuffer io.copyBuffer
func copyBuffer(io.Writer, io.Reader, []byte) (int64, error)
