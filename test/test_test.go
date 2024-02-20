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

package test

import (
	"testing"
)

var config = &Config{
	Print:  true,
	Trace:  true,
	Detail: true,
	Msg:    "%s",
}

func TestAll(t *testing.T) {
	gt := New(t, config)
	gt.Equal(1, 1, "1 == 1")
	gt.NotEqual(1, 2, "1 != 2")
	gt.True(true, "true is true")
	gt.False(false, "false is false")
}

func TestTable(t *testing.T) {
	data := [][]string{
		{"Col1", "Col2", "Col3"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	PrintTable(data, true)
}
