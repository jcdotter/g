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

var Vals map[string]any

type StructBool1 struct{ V1 bool }
type StructInt1 struct{ V1 int }
type StructArray1 struct{ V1 [1]string }
type StructSlice1 struct{ V1 []string }
type StructMap1 struct{ V1 map[string]string }
type StructString1 struct{ V1 string }
type StructStruct1 struct{ V1 StructString1 }
type StructPtrBool1 struct{ V1 *bool }
type StructPtrInt1 struct{ V1 *int }
type StructPtrString1 struct{ V1 *string }
type StructPtrArray1 struct{ V1 *[1]string }
type StructPtrSlice1 struct{ V1 *[]string }
type StructPtrMap1 struct{ V1 *map[string]string }
type StructPtrStruct1 struct{ V1 *StructString1 }
type StructAny1 struct{ V1 any }
type StructBool struct{ V1, V2 bool }
type StructInt struct{ V1, V2 int }
type StructArray struct{ V1, V2 [2]string }
type StructSlice struct{ V1, V2 []string }
type StructMap struct{ V1, V2 map[string]string }
type StructString struct{ V1, V2 string }
type StructStruct struct{ V1, V2 StructString }
type StructPtrBool struct{ V1, V2 *bool }
type StructPtrInt struct{ V1, V2 *int }
type StructPtrString struct{ V1, V2 *string }
type StructPtrArray struct{ V1, V2 *[2]string }
type StructPtrSlice struct{ V1, V2 *[]string }
type StructPtrMap struct{ V1, V2 *map[string]string }
type StructPtrStruct struct{ V1, V2 *StructString }
type StructAny struct{ V1, V2 any }

func init() {
	Vals = createTestVars(false, 0, "false")
}

/* func getTestVarsGmap() Gmap {
	m := MapOf(createTestVars(false, 0, "false")).Gmap()
	m.SortByKeys()
	return m
} */

func createTestVars(b bool, i int, s string, item ...string) map[string]any {
	v := map[string]any{
		"bool":   b,
		"int":    i,
		"string": s,

		"*bool":   &b,
		"*int":    &i,
		"*string": &s,

		"*[1]string":            &[1]string{s},
		"*[]string{1}":          &[]string{s},
		"*map[string]string{1}": &map[string]string{"0": s},
		"*struct(string){1}":    &StructString1{s},

		"*[2]string":            &[2]string{s, s},
		"*[]string{2}":          &[]string{s, s},
		"*map[string]string{2}": &map[string]string{"0": s, "1": s},
		"*struct(string){2}":    &StructString{s, s},

		"[1]bool":                 [1]bool{b},
		"[1]int":                  [1]int{i},
		"[1]string":               [1]string{s},
		"[1][1]string":            [1][1]string{{s}},
		"[1][]string":             [1][]string{{s}},
		"[1]map[string]string{1}": [1]map[string]string{{"0": s}},
		"[1]struct(string){1}":    [1]StructString1{{s}},
		"[1]any(string)":          [1]any{s},

		"[2]bool":                 [2]bool{b, b},
		"[2]int":                  [2]int{i, i},
		"[2]string":               [2]string{s, s},
		"[2][2]string":            [2][2]string{{s, s}, {s, s}},
		"[2][1]string":            [2][1]string{{s}, {s}},
		"[2][]string":             [2][]string{{s, s}, {s, s}},
		"[2]map[string]string{2}": [2]map[string]string{{"0": s, "1": s}, {"0": s, "1": s}},
		"[2]struct(string){2}":    [2]StructString{{s, s}, {s, s}},
		"[2]struct(string){1}":    [2]StructString1{{s}, {s}},
		"[2]any(string)":          [2]any{s, s},

		"[1]*bool":                 [1]*bool{&b},
		"[1]*int":                  [1]*int{&i},
		"[1]*string":               [1]*string{&s},
		"[1]*[1]string":            [1]*[1]string{{s}},
		"[1]*[]string":             [1]*[]string{{s}},
		"[1]*map[string]string{1}": [1]*map[string]string{{"0": s}},
		"[1]*struct(string){1}":    [1]*StructString1{{s}},
		"[1]any(*string)":          [1]any{&s},

		"[2]*bool":                 [2]*bool{&b, &b},
		"[2]*int":                  [2]*int{&i, &i},
		"[2]*string":               [2]*string{&s, &s},
		"[2]*[2]string":            [2]*[2]string{{s, s}, {s, s}},
		"[2]*[1]string":            [2]*[1]string{{s}, {s}},
		"[2]*[]string{2}":          [2]*[]string{{s, s}, {s, s}},
		"[2]*map[string]string{2}": [2]*map[string]string{{"0": s, "1": s}, {"0": s, "1": s}},
		"[2]*struct(string){2}":    [2]*StructString{{s, s}, {s, s}},
		"[2]*struct(string){1}":    [2]*StructString1{{s}, {s}},
		"[2]any(*string)":          [2]any{&s, &s},

		"[]bool{1}":                []bool{b},
		"[]int{1}":                 []int{i},
		"[]string{1}":              []string{s},
		"[][1]string{1}":           [][1]string{{s}},
		"[][]string{1,1}":          [][]string{{s}},
		"[]map[string]string{1,1}": []map[string]string{{"0": s}},
		"[]struct(string){1,1}":    []StructString1{{s}},
		"[]any(string){1}":         []any{s},

		"[]bool{2}":                []bool{b, b},
		"[]int{2}":                 []int{i, i},
		"[]string{2}":              []string{s, s},
		"[][2]string{2}":           [][2]string{{s, s}, {s, s}},
		"[][]string{2,2}":          [][]string{{s, s}, {s, s}},
		"[]map[string]string{2,2}": []map[string]string{{"0": s, "1": s}, {"0": s, "1": s}},
		"[]struct(string){2,2}":    []StructString{{s, s}, {s, s}},
		"[]any(string){2}":         []any{s, s},

		"[]*bool{1}":                []*bool{&b},
		"[]*int{1}":                 []*int{&i},
		"[]*string{1}":              []*string{&s},
		"[]*[1]string{1}":           []*[1]string{{s}},
		"[]*[1]string{2}":           []*[1]string{{s}, {s}},
		"[]*[]string{1,1}":          []*[]string{{s}},
		"[]*map[string]string{1,1}": []*map[string]string{{"0": s}},
		"[]*struct(string){1,1}":    []*StructString1{{s}},
		"[]any(*string){1}":         []any{&s},

		"[]*bool{2}":                []*bool{&b, &b},
		"[]*int{2}":                 []*int{&i, &i},
		"[]*string{2}":              []*string{&s, &s},
		"[]*[2]string{2}":           []*[2]string{{s, s}, {s, s}},
		"[]*[]string{2,2}":          []*[]string{{s, s}, {s, s}},
		"[]*map[string]string{2,2}": []*map[string]string{{"0": s, "1": s}, {"0": s, "1": s}},
		"[]*StructString{2,2}":      []*StructString{{s, s}, {s, s}},
		"[]any(*string){2}":         []any{&s, &s},

		"map[string]bool{1}":                map[string]bool{"0": b},
		"map[string]int{1}":                 map[string]int{"0": i},
		"map[string]string{1}":              map[string]string{"0": s},
		"map[string][1]string{1}":           map[string][1]string{"0": {s}},
		"map[string][]string{1,1}":          map[string][]string{"0": {s}},
		"map[string]map[string]string{1,1}": map[string]map[string]string{"0": {"0": s}},
		"map[string]struct(string){1,1}":    map[string]StructString1{"0": {s}},
		"map[string]any(string){1}":         map[string]any{"0": s},

		"map[string]bool{2}":                map[string]bool{"0": b, "1": b},
		"map[string]int{2}":                 map[string]int{"0": i, "1": i},
		"map[string]string{2}":              map[string]string{"0": s, "1": s},
		"map[string][2]string{2}":           map[string][2]string{"0": {s, s}, "1": {s, s}},
		"map[string][]string{2,2}":          map[string][]string{"0": {s, s}, "1": {s, s}},
		"map[string]map[string]string{2,2}": map[string]map[string]string{"0": {"0": s, "1": s}, "1": {"0": s, "1": s}},
		"map[string]struct(string){2,2}":    map[string]StructString{"0": {s, s}, "1": {s, s}},
		"map[string]any(string){2}":         map[string]any{"0": s, "1": s},

		"map[string]*bool{1}":              map[string]*bool{"0": &b},
		"map[string]*int{1}":               map[string]*int{"0": &i},
		"map[string]*string{1}":            map[string]*string{"0": &s},
		"map[string]*[1]string{1}":         map[string]*[1]string{"0": {s}},
		"map[string]*[]string{1}":          map[string]*[]string{"0": {s}},
		"map[string]*map[string]string{1}": map[string]*map[string]string{"0": {"0": s}},
		"map[string]*struct(string){1}":    map[string]*StructString1{"0": {s}},
		"map[string]any(*string){1}":       map[string]any{"0": &s},

		"map[string]*bool{2}":                map[string]*bool{"0": &b, "1": &b},
		"map[string]*int{2}":                 map[string]*int{"0": &i, "1": &i},
		"map[string]*string{2}":              map[string]*string{"0": &s, "1": &s},
		"map[string]*[2]string{2}":           map[string]*[2]string{"0": {s, s}, "1": {s, s}},
		"map[string]*[]string{2,2}":          map[string]*[]string{"0": {s, s}, "1": {s, s}},
		"map[string]*map[string]string{2,2}": map[string]*map[string]string{"0": {"0": s, "1": s}, "1": {"0": s, "1": s}},
		"map[string]*struct(string){2}":      map[string]*StructString{"0": {s, s}, "1": {s, s}},
		"map[string]any(*string){2}":         map[string]any{"0": &s, "1": &s},

		"struct(bool){1}":                 StructBool1{b},
		"struct(int){1}":                  StructInt1{i},
		"struct(string){1}":               StructString1{s},
		"struct([1]string){1}":            StructArray1{[1]string{s}},
		"struct([]string{1}){1}":          StructSlice1{[]string{s}},
		"struct(map[string]string{1}){1}": StructMap1{map[string]string{"0": s}},
		"struct(struct(string){1}){1}":    StructStruct1{StructString1{s}},
		"struct(any(string)){1}":          StructAny1{s},

		"struct(bool){2}":                 StructBool{b, b},
		"struct(int){2}":                  StructInt{i, i},
		"struct(string){2}":               StructString{s, s},
		"struct([2]string){2}":            StructArray{[2]string{s, s}, [2]string{s, s}},
		"struct([]string{2}){2}":          StructSlice{[]string{s, s}, []string{s, s}},
		"struct(map[string]string{2}){2}": StructMap{map[string]string{"0": s, "1": s}, map[string]string{"0": s, "1": s}},
		"struct(struct(string){2}){2}":    StructStruct{StructString{s, s}, StructString{s, s}},
		"struct(any(string)){2}":          StructAny{s, s},

		"struct(*bool){1}":                 StructPtrBool1{&b},
		"struct(*int){1}":                  StructPtrInt1{&i},
		"struct(*string){1}":               StructPtrString1{&s},
		"struct(*[1]string){1}":            StructPtrArray1{&[1]string{s}},
		"struct(*[]string{1}){1}":          StructPtrSlice1{&[]string{s}},
		"struct(*map[string]string{1}){1}": StructPtrMap1{&map[string]string{"0": s}},
		"struct(*struct(string){1}){1}":    StructPtrStruct1{&StructString1{s}},
		"struct(any(*string)){1}":          StructAny1{&s},

		"struct(*bool){2}":                 StructPtrBool{&b, &b},
		"struct(*int){2}":                  StructPtrInt{&i, &i},
		"struct(*string){2}":               StructPtrString{&s, &s},
		"struct(*[2]string){2}":            StructPtrArray{&[2]string{s, s}, &[2]string{s, s}},
		"struct(*[]string{2}){2}":          StructPtrSlice{&[]string{s, s}, &[]string{s, s}},
		"struct(*map[string]string{2}){2}": StructPtrMap{&map[string]string{"0": s, "1": s}, &map[string]string{"0": s, "1": s}},
		"struct(*struct(string){2}){2}":    StructPtrStruct{&StructString{s, s}, &StructString{s, s}},
		"struct(*any(string)){2}":          StructAny{&s, &s},
	}

	if len(item) > 0 {
		r := map[string]any{}
		for _, v := range item {
			r[v] = v
		}
		return r
	}
	return v
}
