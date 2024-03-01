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

package data

import (
	"sync"
	"unsafe"

	"github.com/jcdotter/go/typ"
)

// -----------------------------------------------------------------------------
// DATA
// Data is a simple data store that holds a list of elements and provides
// methods for adding, removing, and updating elements in the list. It also
// provides a method for getting the index of an element in the list.
//
// In benchmark tests, Data is slower than a slice or map when the len is less
// than 32. However, when the len is greater than 32, Data comperable to a map
// and faster than a slice, while still maintaining the ability to access
// elements by index.

var (
	Cap      = 8
	IndexMin = 32
)

type Data struct {
	sync.Mutex
	k string         // data block identifier
	t uintptr        // data type
	i map[string]int // data index
	l []Elem         // data list
}

type Elem interface {
	Key() string
}

func Make[T Elem](cap int) (d *Data) {
	return maker(*new(T), cap)
}

func maker(T Elem, cap int) (d *Data) {
	d = &Data{
		t: uintptr(unsafe.Pointer(typ.TypeOf(T))),
		l: make([]Elem, 0, cap),
	}
	d.makeIndex(cap)
	return
}

func Of(elems ...Elem) (d *Data) {
	if l := len(elems); l > 0 {
		d = maker(elems[0], max(Cap, l))
		for _, v := range elems {
			d.Add(v)
		}
	}
	return
}

func (d *Data) Key() string {
	return d.k
}

func (d *Data) SetKey(key string) *Data {
	d.k = key
	return d
}

func (d *Data) makeIndex(cap int) {
	if d != nil {
		if d.i != nil || cap < IndexMin {
			return
		}
		d.Lock()
		defer d.Unlock()
		d.i = make(map[string]int, cap)
		for i, v := range d.l {
			d.i[v.Key()] = i
		}
	}
}

func (d *Data) Len() int {
	if d == nil {
		return 0
	}
	return len(d.l)
}

func (d *Data) Valid(value Elem) {
	if d == nil || value == nil {
		return
	}
	if uintptr(unsafe.Pointer(typ.TypeOf(value))) != d.t {
		panic("data: invalid element type")
	}
}

func (d *Data) IndexOf(key string) int {
	if d == nil {
		return -1
	}
	if d.i != nil {
		if i, ok := d.i[key]; ok {
			return i
		}
	} else {
		for i, v := range d.l {
			if v != nil && v.Key() == key {
				return i
			}
		}
	}
	return -1
}

func (d *Data) Has(key string) bool {
	return d.IndexOf(key) != -1
}

func (d *Data) Index(i int) Elem {
	if d == nil {
		return nil
	}
	return d.l[i]
}

func (d *Data) Get(key string) Elem {
	if i := d.IndexOf(key); i > -1 {
		return d.l[i]
	}
	return nil
}

func (d *Data) Add(value Elem) *Data {
	if d == nil {
		d = maker(value, Cap)
	} else {
		if value == nil {
			d.l = append(d.l, nil)
			return d
		}
		d.Valid(value)
		d.makeIndex(len(d.l))
	}
	k := value.Key()
	if i := d.IndexOf(k); i > -1 {
		d.l[i] = value
		return d
	}
	return d.UnsafeAdd(k, value)
}

func (d *Data) UnsafeAdd(key string, value Elem) *Data {
	if d.i != nil {
		d.Lock()
		defer d.Unlock()
		d.i[key] = len(d.l)
	}
	d.l = append(d.l, value)
	return d
}

func (d *Data) SetIndex(i int, value Elem) *Data {
	d.Valid(value)
	if d != nil && i > -1 && i < len(d.l) {
		d.Lock()
		defer d.Unlock()
		if oldKey, newKey := d.l[i].Key(), value.Key(); oldKey != newKey {
			delete(d.i, oldKey)
			d.i[newKey] = i
		}
		d.l[i] = value
	} else {
		panic("data: index out of range")
	}
	return d
}

func (d *Data) Set(key string, value Elem) *Data {
	d.Valid(value)
	if i := d.IndexOf(key); i > -1 {
		d.Lock()
		defer d.Unlock()
		d.l[i] = value
		return d
	}
	return d.UnsafeAdd(key, value)
}

func (d *Data) Remove(name string) *Data {
	if i := d.IndexOf(name); i > -1 {
		d.Lock()
		defer d.Unlock()
		d.l = append(d.l[:i], d.l[i+1:]...)
		delete(d.i, name)
	}
	return d
}

func (d *Data) List() []Elem {
	if d == nil {
		return nil
	}
	return d.l
}
