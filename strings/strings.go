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

package strings

import (
	"strings"
)

// Width returns string s trucated or expanded with whitespaces
// to meet the N number of chars provided
func Width(s string, n int) string {
	l := n - len(s)
	if l > 0 {
		s += strings.Repeat(" ", l)
	} else if l < 0 {
		s = s[:n]
	}
	return s
}
