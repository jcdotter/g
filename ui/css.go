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

package ui

import (
	"strings"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/data"
)

var CSS *data.Data

// -----------------------------------------------------------------------------
// CLASS

// Class represents a CSS class.
type Class struct {
	Sel   string
	Props []string
	Vals  []string
	Block string
}

func Cls(class string) *Class {
	if c := CSS.Get(class); c != nil {
		return c.(*Class)
	}
	var pos int
	for pos > -1 {
		var val string
		if decl, ok := decls[class]; ok {
			c := cls(class, decl(class, val)...)
			CSS.Set(class, c)
			return c
		}
		if pos = strings.LastIndex(class, "-"); pos > -1 {
			class, val = class[:pos], class[pos+1:]
		}
	}
	return nil
}

// Cls creates a new CSS class.
func cls(selector string, propVals ...string) *Class {

	// Create new class
	c := &Class{
		Sel:   selector,
		Props: make([]string, 0, len(propVals)/2),
		Vals:  make([]string, 0, len(propVals)/2),
	}

	// init buffer for class block
	b := buffer.Pool.Get()
	defer buffer.Pool.Put(b)
	b.WriteByte('.')
	b.WriteString(selector)
	b.WriteString(" {")

	// add properties and values to class
	for i := 0; i < len(propVals); i += 2 {
		c.Props = append(c.Props, propVals[i])
		c.Vals = append(c.Vals, propVals[i+1])
		b.WriteString(propVals[i])
		b.WriteByte(':')
		b.WriteString(propVals[i+1])
		b.WriteByte(';')
	}

	// close class block and return
	b.WriteByte('}')
	c.Block = b.String()
	return c
}

// Key returns the class selector.
func (c *Class) Key() string {
	return c.Sel
}

// -----------------------------------------------------------------------------
// CONFIG

// Procedure:
// 1. search created classes for class name
// 2. if not found, parse class name for decls
// 3. if decls found, create class object and add to classes
// 4. return class object

var decls = map[string]decl{
	"aspect":         css_aspect,
	"container":      css_container,
	"columns":        css_columns,
	"break-after":    css_break,
	"break-before":   css_break,
	"break-inside":   css_break,
	"box-decoration": css_box,
	"box-border":     css_box,
	"box-content":    css_box,
	"block":          css_display,
	"flex":           css_display,
	"grid":           css_display,
	"table":          css_display,
	"flow-root":      css_display,
	"contents":       css_display,
	"list-item":      css_display,
	"inline":         css_display,
	"hidden":         css_display,
}

var css_size = map[string]string{
	"auto": "auto",
	"px":   "1px",
	"3xs":  "16rem",
	"2xs":  "18rem",
	"xs":   "20rem",
	"sm":   "24rem",
	"md":   "28rem",
	"lg":   "32rem",
	"xl":   "36rem",
	"2xl":  "42rem",
	"3xl":  "48rem",
	"4xl":  "56rem",
	"5xl":  "64rem",
	"6xl":  "72rem",
	"7xl":  "80rem",
	"0":    "0",
	"1":    "0.25rem",
	"2":    "0.5rem",
	"3":    "0.75rem",
	"4":    "1rem",
	"5":    "1.25rem",
	"6":    "1.5rem",
	"7":    "1.75rem",
	"8":    "2rem",
	"9":    "2.25rem",
	"10":   "2.5rem",
	"12":   "3rem",
	"16":   "4rem",
	"20":   "5rem",
	"24":   "6rem",
	"28":   "7rem",
	"32":   "8rem",
	"36":   "9rem",
	"40":   "10rem",
	"44":   "11em",
	"48":   "12rem",
	"52":   "13rem",
	"56":   "14rem",
	"60":   "15rem",
	"64":   "16rem",
	"72":   "18rem",
	"80":   "20rem",
	"96":   "24rem",
}

type decl func(string, string) []string

func css_aspect(c string, v string) []string {
	switch v {
	case "auto":
	case "square":
		v = "1/1"
	case "video":
		v = "16/9"
	default:
		if v = parseBracket(v); v == "" {
			return nil
		}
	}
	return []string{"aspect-ratio", v}
}

func css_container(c string, v string) []string {
	return []string{"width", "100%"}
}

func css_columns(c string, v string) []string {
	switch v {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "12", "auto":
	default:
		if n, ok := css_size[v]; ok {
			v = n
		} else if v = parseBracket(v); v == "" {
			return nil
		}
	}
	return []string{"columns", v}
}

func css_break(c string, v string) []string {
	switch v {
	case "auto", "always", "all", "column", "page", "left", "right", "avoid", "avoid-page", "avoid-column":
	default:
		return nil
	}
	return []string{c, v}
}

func css_box(c string, v string) []string {
	switch v {
	case "", "clone", "slice":
	default:
		return nil
	}
	switch c {
	case "box-decoration":
		c = "box-decoration-break"
	case "box-border", "box-content":
		v = c
		c = "box-sizing"
	default:
		return nil
	}
	return []string{c, v}
}

func css_display(c string, v string) []string {
	switch v {
	case "":
		v = c
	case "cell", "row", "column", "column-group", "footer-group", "header-group", "row-group":
		if c != "table" {
			return nil
		}
	case "block", "flex", "grid", "table":
		if c != "inline" {
			return nil
		}
	default:
		return nil
	}
	return []string{"display", v}
}

func parseBracket(v string) string {
	if v[0] == '[' && v[len(v)-1] == ']' {
		return v[1 : len(v)-1]
	}
	return ""
}

// -----------------------------------------------------------------------------
// CLASSES

// Classes is a collection of CSS class utilities.
// The class names and their properties are based on
// the Tailwind CSS framework.
var Classes = data.Of(

	// -------------------------------------------------------------------------
	// LAYOUT

	// aspect ratio
	cls("aspect-auto", "aspect-ratio", "auto"),
	cls("aspect-square", "aspect-ratio", "1/1"),
	cls("aspect-video", "aspect-ratio", "16/9"),
	cls("aspect-16/9", "aspect-ratio", "16/9"),
	cls("aspect-4/3", "aspect-ratio", "4/3"),

	// container
	cls("container", "width", "100%"),

	// columns
	cls("columns-1", "columns", "1"),
	cls("columns-2", "columns", "2"),
	cls("columns-3", "columns", "3"),
	cls("columns-4", "columns", "4"),
	cls("columns-5", "columns", "5"),
	cls("columns-6", "columns", "6"),
	cls("columns-7", "columns", "7"),
	cls("columns-8", "columns", "8"),
	cls("columns-9", "columns", "9"),
	cls("columns-10", "columns", "10"),
	cls("columns-12", "columns", "12"),
	cls("columns-auto", "columns", "auto"),
	cls("columns-3xs", "columns", "16rem"),
	cls("columns-2xs", "columns", "18rem"),
	cls("columns-xs", "columns", "20rem"),
	cls("columns-sm", "columns", "24rem"),
	cls("columns-md", "columns", "28rem"),
	cls("columns-lg", "columns", "32rem"),
	cls("columns-xl", "columns", "36rem"),
	cls("columns-2xl", "columns", "42rem"),
	cls("columns-3xl", "columns", "48rem"),
	cls("columns-4xl", "columns", "56rem"),
	cls("columns-5xl", "columns", "64rem"),
	cls("columns-6xl", "columns", "72rem"),
	cls("columns-7xl", "columns", "80rem"),

	// break
	cls("break-after-auto", "break-after", "auto"),
	cls("break-after-avoid", "break-after", "avoid"),
	cls("break-after-all", "break-after", "all"),
	cls("break-after-avoid-page", "break-after", "avoid-page"),
	cls("break-after-page", "break-after", "page"),
	cls("break-after-left", "break-after", "left"),
	cls("break-after-right", "break-after", "right"),
	cls("break-after-column", "break-after", "column"),
	cls("break-before-auto", "break-before", "auto"),
	cls("break-before-avoid", "break-before", "avoid"),
	cls("break-before-all", "break-before", "all"),
	cls("break-before-avoid-page", "break-before", "avoid-page"),
	cls("break-before-page", "break-before", "page"),
	cls("break-before-left", "break-before", "left"),
	cls("break-before-right", "break-before", "right"),
	cls("break-before-column", "break-before", "column"),
	cls("break-inside-auto", "break-inside", "auto"),
	cls("break-inside-avoid", "break-inside", "avoid"),
	cls("break-inside-avoid-page", "break-inside", "avoid-page"),
	cls("break-inside-avoid-column", "break-inside", "avoid-column"),

	// box
	cls("box-decoration-clone", "box-decoration-break", "clone"),
	cls("box-decoration-slice", "box-decoration-break", "slice"),
	cls("box-border", "box-sizing", "border-box"),
	cls("box-content", "box-sizing", "content-box"),

	// display
	cls("block", "display", "block"),
	cls("flex", "display", "flex"),
	cls("grid", "display", "grid"),
	cls("table", "display", "table"),
	cls("table-cell", "display", "table-cell"),
	cls("table-row", "display", "table-row"),
	cls("table-column", "display", "table-column"),
	cls("table-caption", "display", "table-caption"),
	cls("table-column-group", "display", "table-column-group"),
	cls("table-header-group", "display", "table-header-group"),
	cls("table-footer-group", "display", "table-footer-group"),
	cls("flow-root", "display", "flow-root"),
	cls("contents", "display", "contents"),
	cls("list-item", "display", "list-item"),
	cls("inline", "display", "inline"),
	cls("inline-block", "display", "inline-block"),
	cls("inline-flex", "display", "inline-flex"),
	cls("inline-grid", "display", "inline-grid"),
	cls("inline-table", "display", "inline-table"),
	cls("hidden", "display", "none"),

	// float
	cls("float-start", "float", "inline-start"),
	cls("float-end", "float", "inline-end"),
	cls("float-right", "float", "right"),
	cls("float-left", "float", "left"),
	cls("float-none", "float", "none"),

	// clear
	cls("clear-start", "clear", "inline-start"),
	cls("clear-end", "clear", "inline-end"),
	cls("clear-left", "clear", "left"),
	cls("clear-right", "clear", "right"),
	cls("clear-both", "clear", "both"),
	cls("clear-none", "clear", "none"),

	// isolation
	cls("isolate", "isolation", "isolate"),
	cls("isolate-auto", "isolation", "auto"),

	// object
	cls("object-contain", "object-fit", "contain"),
	cls("object-cover", "object-fit", "cover"),
	cls("object-fill", "object-fit", "fill"),
	cls("object-none", "object-fit", "none"),
	cls("object-scale-down", "object-fit", "scale-down"),
	cls("object-bottom", "object-position", "bottom"),
	cls("object-center", "object-position", "center"),
	cls("object-left", "object-position", "left"),
	cls("object-left-bottom", "object-position", "left bottom"),
	cls("object-left-top", "object-position", "left top"),
	cls("object-right", "object-position", "right"),
	cls("object-right-bottom", "object-position", "right bottom"),
	cls("object-right-top", "object-position", "right top"),
	cls("object-top", "object-position", "top"),

	// overflow
	cls("overflow-auto", "overflow", "auto"),
	cls("overflow-hidden", "overflow", "hidden"),
	cls("overflow-clip", "overflow", "clip"),
	cls("overflow-visible", "overflow", "visible"),
	cls("overflow-scroll", "overflow", "scroll"),
	cls("overflow-x-auto", "overflow-x", "auto"),
	cls("overflow-x-hidden", "overflow-x", "hidden"),
	cls("overflow-x-clip", "overflow-x", "clip"),
	cls("overflow-x-visible", "overflow-x", "visible"),
	cls("overflow-x-scroll", "overflow-x", "scroll"),
	cls("overflow-y-auto", "overflow-y", "auto"),
	cls("overflow-y-hidden", "overflow-y", "hidden"),
	cls("overflow-y-clip", "overflow-y", "clip"),
	cls("overflow-y-visible", "overflow-y", "visible"),
	cls("overflow-y-scroll", "overflow-y", "scroll"),

	// overscroll
	cls("overscroll-auto", "overscroll-behavior", "auto"),
	cls("overscroll-contain", "overscroll-behavior", "contain"),
	cls("overscroll-none", "overscroll-behavior", "none"),
	cls("overscroll-y-auto", "overscroll-behavior-y", "auto"),
	cls("overscroll-y-contain", "overscroll-behavior-y", "contain"),
	cls("overscroll-y-none", "overscroll-behavior-y", "none"),
	cls("overscroll-x-auto", "overscroll-behavior-x", "auto"),
	cls("overscroll-x-contain", "overscroll-behavior-x", "contain"),
	cls("overscroll-x-none", "overscroll-behavior-x", "none"),

	// position
	cls("static", "position", "static"),
	cls("fixed", "position", "fixed"),
	cls("absolute", "position", "absolute"),
	cls("relative", "position", "relative"),
	cls("sticky", "position", "sticky"),

	// inset
	cls("inset-0", "top", "0", "right", "0", "bottom", "0", "left", "0"),
	cls("inset-auto", "top", "auto", "right", "auto", "bottom", "auto", "left", "auto"),
	cls("inset-y-0", "top", "0", "bottom", "0"),
	cls("inset-x-0", "right", "0", "left", "0"),
	cls("inset-y-auto", "top", "auto", "bottom", "auto"),
	cls("inset-x-auto", "right", "auto", "left", "auto"),
	// TODO: fill in inset classes

	// visibility
	cls("visible", "visibility", "visible"),
	cls("invisible", "visibility", "hidden"),
	cls("collapse", "visibility", "collapse"),

	// z-index
	cls("z-0", "z-index", "0"),
	cls("z-10", "z-index", "10"),
	cls("z-20", "z-index", "20"),
	cls("z-30", "z-index", "30"),
	cls("z-40", "z-index", "40"),
	cls("z-50", "z-index", "50"),
	cls("z-auto", "z-index", "auto"),

	// -------------------------------------------------------------------------
	// FLEX & GRID

	// basis

	// flex
	cls("flex-row", "flex-direction", "row"),
	cls("flex-row-reverse", "flex-direction", "row-reverse"),
	cls("flex-col", "flex-direction", "column"),
	cls("flex-col-reverse", "flex-direction", "column-reverse"),
	cls("flex-wrap", "flex-wrap", "wrap"),
	cls("flex-wrap-reverse", "flex-wrap", "wrap-reverse"),
	cls("flex-nowrap", "flex-wrap", "nowrap"),
	cls("flex-1", "flex", "1 1 0%"),
	cls("flex-auto", "flex", "1 1 auto"),
	cls("flex-initial", "flex", "0 1 auto"),
	cls("flex-none", "flex", "none"),

	// flex-grow
	cls("flex-grow-0", "flex-grow", "0"),
	cls("flex-grow", "flex-grow", "1"),

	// flex-shrink
	cls("flex-shrink-0", "flex-shrink", "0"),
	cls("flex-shrink", "flex-shrink", "1"),

	// order
	cls("order-1", "order", "1"),
	cls("order-2", "order", "2"),
	cls("order-3", "order", "3"),
	cls("order-4", "order", "4"),
	cls("order-5", "order", "5"),
	cls("order-6", "order", "6"),
	cls("order-7", "order", "7"),
	cls("order-8", "order", "8"),
	cls("order-9", "order", "9"),
	cls("order-10", "order", "10"),
	cls("order-11", "order", "11"),
	cls("order-12", "order", "12"),
	cls("order-first", "order", "-9999"),
	cls("order-last", "order", "9999"),
	cls("order-none", "order", "0"),

	// grid
	cls("grid-cols-1", "grid-template-columns", "repeat(1, minmax(0, 1fr))"),
	cls("grid-cols-2", "grid-template-columns", "repeat(2, minmax(0, 1fr))"),
	cls("grid-cols-3", "grid-template-columns", "repeat(3, minmax(0, 1fr))"),
	cls("grid-cols-4", "grid-template-columns", "repeat(4, minmax(0, 1fr))"),
	cls("grid-cols-5", "grid-template-columns", "repeat(5, minmax(0, 1fr))"),
	cls("grid-cols-6", "grid-template-columns", "repeat(6, minmax(0, 1fr))"),
	cls("grid-cols-7", "grid-template-columns", "repeat(7, minmax(0, 1fr))"),
	cls("grid-cols-8", "grid-template-columns", "repeat(8, minmax(0, 1fr))"),
	cls("grid-cols-9", "grid-template-columns", "repeat(9, minmax(0, 1fr))"),
	cls("grid-cols-10", "grid-template-columns", "repeat(10, minmax(0, 1fr))"),
	cls("grid-cols-11", "grid-template-columns", "repeat(11, minmax(0, 1fr))"),
	cls("grid-cols-12", "grid-template-columns", "repeat(12, minmax(0, 1fr))"),
	cls("grid-cols-none", "grid-template-columns", "none"),
	cls("grid-cols-subgrid", "grid-template-columns", "subgrid"),
	cls("grid-rows-1", "grid-template-rows", "repeat(1, minmax(0, 1fr))"),
	cls("grid-rows-2", "grid-template-rows", "repeat(2, minmax(0, 1fr))"),
	cls("grid-rows-3", "grid-template-rows", "repeat(3, minmax(0, 1fr))"),
	cls("grid-rows-4", "grid-template-rows", "repeat(4, minmax(0, 1fr))"),
	cls("grid-rows-5", "grid-template-rows", "repeat(5, minmax(0, 1fr))"),
	cls("grid-rows-6", "grid-template-rows", "repeat(6, minmax(0, 1fr))"),
	cls("grid-rows-7", "grid-template-rows", "repeat(7, minmax(0, 1fr))"),
	cls("grid-rows-8", "grid-template-rows", "repeat(8, minmax(0, 1fr))"),
	cls("grid-rows-9", "grid-template-rows", "repeat(9, minmax(0, 1fr))"),
	cls("grid-rows-10", "grid-template-rows", "repeat(10, minmax(0, 1fr))"),
	cls("grid-rows-11", "grid-template-rows", "repeat(11, minmax(0, 1fr))"),
	cls("grid-rows-12", "grid-template-rows", "repeat(12, minmax(0, 1fr))"),
	cls("grid-rows-none", "grid-template-rows", "none"),
	cls("grid-rows-subgrid", "grid-template-rows", "subgrid"),
	cls("gap-0", "gap", "0"),
	cls("gap-1", "gap", "0.25rem"),
	cls("gap-2", "gap", "0.5rem"),
	cls("gap-3", "gap", "0.75rem"),
	cls("gap-4", "gap", "1rem"),
	cls("gap-5", "gap", "1.25rem"),
	cls("gap-6", "gap", "1.5rem"),
	cls("gap-7", "gap", "1.75rem"),
	cls("gap-8", "gap", "2rem"),
	cls("gap-9", "gap", "2.25rem"),
	cls("gap-10", "gap", "2.5rem"),
	cls("gap-12", "gap", "3rem"),
	cls("gap-16", "gap", "4rem"),
	cls("gap-20", "gap", "5rem"),
	cls("gap-24", "gap", "6rem"),
	cls("gap-28", "gap", "7rem"),
	cls("gap-32", "gap", "8rem"),
	cls("gap-36", "gap", "9rem"),
	cls("gap-40", "gap", "10rem"),
	cls("gap-44", "gap", "11rem"),
	cls("gap-48", "gap", "12rem"),
	cls("gap-52", "gap", "13rem"),
	cls("gap-56", "gap", "14rem"),
	cls("gap-60", "gap", "15rem"),
	cls("gap-64", "gap", "16rem"),
	cls("gap-72", "gap", "18rem"),
	cls("gap-80", "gap", "20rem"),
	cls("gap-96", "gap", "24rem"),
	cls("gap-px", "gap", "1px"),
	cls("gap-x-0", "column-gap", "0"),
	cls("gap-x-1", "column-gap", "0.25rem"),
	cls("gap-x-2", "column-gap", "0.5rem"),
	cls("gap-x-3", "column-gap", "0.75rem"),
	cls("gap-x-4", "column-gap", "1rem"),
	cls("gap-x-5", "column-gap", "1.25rem"),
	cls("gap-x-6", "column-gap", "1.5rem"),
	cls("gap-x-7", "column-gap", "1.75rem"),
	cls("gap-x-8", "column-gap", "2rem"),
	cls("gap-x-9", "column-gap", "2.25rem"),
	cls("gap-x-10", "column-gap", "2.5rem"),
	cls("gap-x-12", "column-gap", "3rem"),
	cls("gap-x-16", "column-gap", "4rem"),
	cls("gap-x-20", "column-gap", "5rem"),
	cls("gap-x-24", "column-gap", "6rem"),
	cls("gap-x-28", "column-gap", "7rem"),
	cls("gap-x-32", "column-gap", "8rem"),
	cls("gap-x-36", "column-gap", "9rem"),
	cls("gap-x-40", "column-gap", "10rem"),
	cls("gap-x-44", "column-gap", "11rem"),
	cls("gap-x-48", "column-gap", "12rem"),
	cls("gap-x-52", "column-gap", "13rem"),
	cls("gap-x-56", "column-gap", "14rem"),
	cls("gap-x-60", "column-gap", "15rem"),
	cls("gap-x-64", "column-gap", "16rem"),
	cls("gap-x-72", "column-gap", "18rem"),
	cls("gap-x-80", "column-gap", "20rem"),
	cls("gap-x-96", "column-gap", "24rem"),
	cls("gap-x-px", "column-gap", "1px"),
	cls("gap-y-0", "row-gap", "0"),
	cls("gap-y-1", "row-gap", "0.25rem"),
	cls("gap-y-2", "row-gap", "0.5rem"),
	cls("gap-y-3", "row-gap", "0.75rem"),
	cls("gap-y-4", "row-gap", "1rem"),
	cls("gap-y-5", "row-gap", "1.25rem"),
	cls("gap-y-6", "row-gap", "1.5rem"),
	cls("gap-y-7", "row-gap", "1.75rem"),
	cls("gap-y-8", "row-gap", "2rem"),
	cls("gap-y-9", "row-gap", "2.25rem"),
	cls("gap-y-10", "row-gap", "2.5rem"),
	cls("gap-y-12", "row-gap", "3rem"),
	cls("gap-y-16", "row-gap", "4rem"),
	cls("gap-y-20", "row-gap", "5rem"),
	cls("gap-y-24", "row-gap", "6rem"),
	cls("gap-y-28", "row-gap", "7rem"),
	cls("gap-y-32", "row-gap", "8rem"),
	cls("gap-y-36", "row-gap", "9rem"),
	cls("gap-y-40", "row-gap", "10rem"),
	cls("gap-y-44", "row-gap", "11rem"),
	cls("gap-y-48", "row-gap", "12rem"),
	cls("gap-y-52", "row-gap", "13rem"),
	cls("gap-y-56", "row-gap", "14rem"),
	cls("gap-y-60", "row-gap", "15rem"),
	cls("gap-y-64", "row-gap", "16rem"),
	cls("gap-y-72", "row-gap", "18rem"),
	cls("gap-y-80", "row-gap", "20rem"),
	cls("gap-y-96", "row-gap", "24rem"),
	cls("gap-y-px", "row-gap", "1px"),
)

// -----------------------------------------------------------------------------
// CLASS ROUTES

// ClassRoutes returns the routes for the classes.
func ClassRoutes(class string) {
	switch {
	case class[:5] == "aspect":

	}
}
