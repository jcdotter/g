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
	"strconv"
	"testing"

	"github.com/jcdotter/go/test"
)

var config = &test.Config{
	PrintTest:   true,
	PrintFail:   true,
	PrintTrace:  true,
	PrintDetail: true,
	FailFatal:   true,
	Msg:         "%s",
}

type Entities []*Entity
type Entity struct {
	name string
}

func (e *Entity) Key() string {
	return e.name
}

func (e *Entities) Len() int {
	return len(*e)
}

func (e *Entities) Get(name string) *Entity {
	if e != nil {
		for _, ent := range *e {
			if ent.Key() == name {
				return ent
			}
		}
	}
	return nil
}

func (e *Entities) Has(name string) bool {
	return e.Get(name) != nil
}

func (e *Entities) Add(ent *Entity) {
	if ent != nil && !e.Has(ent.Key()) {
		if e == nil {
			*e = make([]*Entity, 0, 8)
		}
		*e = append(*e, ent)
	}
}

func (e *Entities) Remove(name string) {
	if e != nil {
		for i, ent := range *e {
			if ent.Key() == name {
				*e = append((*e)[:i], (*e)[i+1:]...)
				break
			}
		}
	}
}

func TestData(t *testing.T) {
	gt := test.New(t, config)
	gt.Msg = "Data.%s"
	type Struct struct {
		Data *Data
	}
	s := &Struct{Data: Make[*Entity](8)}
	s.Data.Add(&Entity{name: "entity1"})
	s.Data.Add(&Entity{name: "entity2"})
	s.Data.Add(&Entity{name: "entity3"})

	gt.Equal(3, s.Data.Len(), "Len()")
	gt.Equal(0, s.Data.IndexOf("entity1"), "IndexOf(entity1)")
	gt.Equal(1, s.Data.IndexOf("entity2"), "IndexOf(entity2)")
	gt.Equal(2, s.Data.IndexOf("entity3"), "IndexOf(entity3)")
	gt.Equal(-1, s.Data.IndexOf("entity4"), "IndexOf(entity4)")
	gt.True(s.Data.Has("entity1"), "Has(entity1)")
	gt.True(s.Data.Has("entity2"), "Has(entity2)")
	gt.True(s.Data.Has("entity3"), "Has(entity3)")
	gt.False(s.Data.Has("entity4"), "Has(entity4)")
	gt.Equal("entity1", s.Data.Index(0).Key(), "Index(0)")
	gt.Equal("entity2", s.Data.Index(1).Key(), "Index(1)")
	gt.Equal("entity3", s.Data.Index(2).Key(), "Index(2)")
	gt.Equal(nil, s.Data.Get("entity4"), "Get(entity4)")
	gt.Equal(2, s.Data.Remove("entity2").Len(), "Remove(entity).Len()")
	gt.Equal(0, s.Data.IndexOf("entity1"), "IndexOf(entity1)")
	gt.Equal(-1, s.Data.IndexOf("entity2"), "IndexOf(entity2)")
	gt.Equal(1, s.Data.IndexOf("entity3"), "IndexOf(entity3)")

	s.Data.Add(nil)
	gt.Equal(3, s.Data.Len(), "Add(nil).Len()")
	gt.Equal(nil, s.Data.Index(2), "Add(nil).Index()")
	gt.Equal(nil, s.Data.Get("entity4"), "Add(nil).Get()")

}

func BenchmarkAdd(b *testing.B) {
	for _, n := range []int{4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048} {
		sn := strconv.Itoa(n)
		b.Run("map("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m := make(map[string]Entity, 8)
				for j := 0; j < n; j++ {
					name := "entity" + strconv.Itoa(j)
					m[name] = Entity{name: name}
				}
			}
		})
		b.Run("Entity("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e := make(Entities, 0, 8)
				for j := 0; j < n; j++ {
					e.Add(&Entity{name: "entity" + strconv.Itoa(j)})
				}
			}
		})
		b.Run("Data("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				d := Make[*Entity](n)
				for j := 0; j < n; j++ {
					d.Add(&Entity{name: "entity" + strconv.Itoa(j)})
				}
			}
		})
	}
}

func BenchmarkGet(b *testing.B) {
	for _, n := range []int{4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048} {
		sn := strconv.Itoa(n)
		m := make(map[string]Entity, 8)
		for j := 0; j < n; j++ {
			name := "entity" + strconv.Itoa(j)
			m[name] = Entity{name: name}
		}
		b.Run("map("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					name := "entity" + strconv.Itoa(j)
					_ = m[name]
				}
			}
		})
		e := make(Entities, 0, 8)
		for j := 0; j < n; j++ {
			e.Add(&Entity{name: "entity" + strconv.Itoa(j)})
		}
		b.Run("Entity("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					e.Get("entity" + strconv.Itoa(j))
				}
			}
		})
		d := Make[*Entity](n)
		for j := 0; j < n; j++ {
			d.Add(&Entity{name: "entity" + strconv.Itoa(j)})
		}
		b.Run("Data("+sn+")", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					d.Get("entity" + strconv.Itoa(j))
				}
			}
		})
	}
}
