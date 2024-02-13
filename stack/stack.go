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
	"runtime"
)

const (
	stackDepth = 1
	stackSkip  = 2
)

// Stack is a stack of program counters
// used to identify the caller of a log
type Stack struct {
	pc []uintptr
	p  *StackPool
}

// NewStack returns a new stack of the specified depth
func NewStack(depth int) *Stack {
	return &Stack{pc: make([]uintptr, depth)}
}

// Reset empties the stack and keeps the capacity
func (s *Stack) Reset() {
	s.pc[0] = 0
}

// Free returns the stack to the stack pool
func (s *Stack) Free() {
	s.p.Put(s)
}

// Populate populates the stack with the current
// program counters beginning at the specified skip
func (s *Stack) Populate(skip int) *Stack {
	runtime.Callers(skip+stackSkip, s.pc)
	return s
}
