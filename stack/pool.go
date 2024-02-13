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

package stack

import (
	"sync"
)

var Pool = NewStackPool()

// StackPool is a type-safe wrapper around a sync.Pool.
// It is used to pool Stack objects.
type StackPool struct {
	p *sync.Pool
}

// NewStackPool constructs a new StackPool.
func NewStackPool() *StackPool {
	return &StackPool{
		p: &sync.Pool{
			New: func() any { return NewStack(stackDepth) },
		},
	}
}

// Get retrieves a Stack from the pool, creating one if necessary.
func (p *StackPool) Get() *Stack {
	s := p.p.Get().(*Stack)
	s.p = p
	return s
}

// Put returns a Stack to the pool.
func (p *StackPool) Put(c *Stack) {
	c.Reset()
	p.p.Put(c)
}
