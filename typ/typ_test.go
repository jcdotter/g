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
	"fmt"
	"testing"
)

func TestTest(t *testing.T) {
	s := []string{"a", "b", "c"}
	SliceOf(&s).Extend(3)
	s[5] = "f"
	fmt.Println(len(s), s)

}

func TestUid(t *testing.T) {
	u := generateUid()
	u1 := generateUid()

	fmt.Println(u.String())
	fmt.Println(u1.String())
}

func BenchmarkUid(b *testing.B) {
	var u UID
	b.Run("generate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			u = generateUid()
		}
	})
	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = u.String()
		}
	})
}
