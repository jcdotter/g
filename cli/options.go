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

package cli

import (
	"strings"
)

var (
	optDisplayMax   = 7
	optDefault      = []byte(`  `)
	optCUR          = []byte(`> `)
	optCheck        = []byte(`  `)
	optHoverStyle   = Styl(Bold, HiWhite).Codes()
	optChecked      = []byte(`√ `)
	optCheckStyle   = Styl(Faint, White).Codes()
	optCheckedStyle = Styl(Bold, Cyan).Codes()
	optDefaultStyle = Styl(Faint, White).Codes()
	optInstStyle    = instStyl.Codes()
	optFilterInst   = "begin typing to search for values..."
	optFilterRes    = "showing search results for: '"
)

type Options struct {
	Cache      *Message // the cached rendered options
	Opts       []string // the options to display
	Checkboxes bool     // whether to display checkboxes
	Display    int      // the number of options displayed
	Hovered    int      // the index of the selected option
	Selected   []int    // the index of selected options
	Style      OptStyle // the style of the options
	Filter     string   // the filter string for the options displayed
	Index      []int    // the index of the filtered options
}

type OptStyle struct {
	Default  []byte // the default option style codes
	Hover    []byte // the style codes for the hovered option
	Cursor   []byte // the stylized cursor
	NoCursor []byte // the stylized nocursor
	Check    []byte // the stylized checkbox
	Checked  []byte // the stylized checkedbox
}

// NewOptions returns a new Options
func NewOptions(opts []string, check bool) *Options {
	o := &Options{
		Cache:      Msg(),
		Opts:       opts,
		Checkboxes: check,
		Display:    optDisplayMax,
		Index:      make([]int, 0, len(opts)),
		Style: OptStyle{
			Default:  optDefaultStyle,
			Hover:    optHoverStyle,
			Cursor:   append(optHoverStyle, optCUR...),
			NoCursor: append(optDefaultStyle, optDefault...),
			Check:    append(optCheckStyle, optCheck...),
			Checked:  append(optCheckedStyle, optChecked...),
		},
	}
	o.index()
	return o
}

// Select returns the selected options
func (o *Options) Select() []string {
	if !o.Checkboxes {
		return []string{o.Opts[o.Hovered]}
	}
	if len(o.Selected) == 0 {
		return nil
	}
	s := make([]string, len(o.Selected))
	for i, idx := range o.Selected {
		s[i] = o.Opts[idx]
	}
	return s
}

// Render renders the options as a Message for display
func (o *Options) Render() *Message {
	o.Cache.Reset()
	o.Cache.buf.MustWrite(EOL)
	o.Cache.dem = append(o.Cache.dem, 0)
	o.bufferFilter()
	for _, optIdx := range o.truncate() {
		o.bufferOption(optIdx)
		o.Cache.dem = append(o.Cache.dem, len(o.Opts[optIdx]))
	}
	o.Cache.buf.MustWrite(Reset.Code())
	o.Cache.prepared = true
	return o.Cache
}

// bufferOption renders a single option to the buffer
func (o *Options) bufferOption(i int) {
	if i > -1 && i < len(o.Opts) {
		if i == o.Hovered {
			o.Cache.buf.MustWrite(o.Style.Cursor)
			o.bufferCheckbox(i)
			o.Cache.buf.MustWrite(o.Style.Hover)
		} else {
			o.Cache.buf.MustWrite(o.Style.NoCursor)
			o.bufferCheckbox(i)
			o.Cache.buf.MustWrite(o.Style.Default)
		}
		o.Cache.buf.MustWriteString(o.Opts[i]).
			MustWrite(EOL)
		return
	}
	panic(ErrRange)
}

// bufferCheckbox renders a checkbox to the buffer
func (o *Options) bufferCheckbox(i int) {
	if o.Checkboxes {
		if o.IsSelected(i) != -1 {
			o.Cache.buf.MustWrite(o.Style.Checked)
		} else {
			o.Cache.buf.MustWrite(o.Style.Check)
		}
	}
}

// bufferFilter renders the filter to the buffer
func (o *Options) bufferFilter() {
	if len(o.Opts) > optDisplayMax {
		o.Cache.dem = append(o.Cache.dem, make([]int, 2)...)
		o.Cache.buf.MustWrite(optInstStyle)
		if len(o.Filter) > 0 {
			o.Cache.buf.MustWriteString(optFilterRes).
				MustWriteString(o.Filter).
				MustWriteByte('\'')
			o.Cache.dem[len(o.Cache.dem)-2] = len(o.Filter) + len(optFilterRes) + 1
		} else {
			o.Cache.buf.MustWriteString(optFilterInst)
			o.Cache.dem[len(o.Cache.dem)-2] = len(optFilterInst)
		}
		o.Cache.buf.MustWrite(EOL).MustWrite(EOL)
	}
}

// bufferLine

// SetDefault sets the default style
func (o *Options) SetDefault(styles ...Style) {
	o.Style.Default = Styl(styles...).Codes()
	o.Style.NoCursor = Styl(styles...).Msg(optDefault).Bytes()
}

// SetHover sets the hovered style
func (o *Options) SetHover(styles ...Style) {
	o.Style.Hover = Styl(styles...).Codes()
	o.Style.Cursor = Styl(styles...).Msg(optCUR).Bytes()
}

// SetCheckbox sets the checkbox style
func (o *Options) SetCheckbox(styles ...Style) {
	o.Style.Check = Styl(styles...).Msg(optCheck).Bytes()
}

// SetCheckedbox sets the checkedbox style
func (o *Options) SetCheckedbox(styles ...Style) {
	o.Style.Checked = Styl(styles...).Msg(optChecked).Bytes()
}

// Down moves the cursor down one option
func (o *Options) Down() *Message {
	if i := o.HoveredIndex(); i < len(o.Index)-1 {
		o.Hovered = o.Index[i+1]
		return o.Render()
	}
	return o.Cache
}

// Up moves the cursor up one option
func (o *Options) Up() *Message {
	if i := o.HoveredIndex(); i > 0 {
		o.Hovered = o.Index[i-1]
		return o.Render()
	}
	return o.Cache
}

// PageDown moves the cursor down one page
func (o *Options) PageDown() *Message {
	i := o.HoveredIndex()
	if d := (len(o.Index) - 1) - i; d > 0 {
		if d > optDisplayMax {
			d = optDisplayMax
		}
		o.Hovered = o.Index[i+d]
		return o.Render()
	}
	return o.Cache
}

// PageUp moves the cursor up one page
func (o *Options) PageUp() *Message {
	if i := o.HoveredIndex(); i > 0 {
		if d := i - optDisplayMax; d > 0 {
			o.Hovered = o.Index[d]
		} else {
			o.Hovered = o.Index[0]
		}
		return o.Render()
	}
	return o.Cache
}

// Last moves the cursor to the last option
func (o *Options) Last() *Message {
	if i, l := o.HoveredIndex(), len(o.Index)-1; i < l {
		o.Hovered = o.Index[l]
		return o.Render()
	}
	return o.Cache
}

// First moves the cursor to the first option
func (o *Options) First() *Message {
	if i := o.HoveredIndex(); i > 0 {
		o.Hovered = o.Index[0]
		return o.Render()
	}
	return o.Cache
}

// HoveredIndex returns the index of the hovered option
// in the index list
func (o *Options) HoveredIndex() int {
	for i, idx := range o.Index {
		if idx == o.Hovered {
			return i
		}
	}
	o.Hovered = o.Index[0]
	return 0
}

// IsSelected returns the items index in the Selected slice,
// or -1 if it is not selected
func (o *Options) IsSelected(i int) int {
	if o.Selected == nil {
		return -1
	}
	for p, x := range o.Selected {
		if x == i {
			return p
		}
	}
	return -1
}

// Toggle toggles the option selection at the Hovered index
func (o *Options) Toggle() *Message {
	if o.Checkboxes {
		o.toggle()
		return o.Render()
	}
	return o.Cache
}

// toggle toggles the option selection at the Hovered index
func (o *Options) toggle() {
	if i := o.IsSelected(o.Hovered); i != -1 {
		o.Selected = append(o.Selected[:i], o.Selected[i+1:]...)
		return
	}
	o.Selected = append(o.Selected, o.Hovered)
}

// Search updates the filter with the given character and renders the options
func (o *Options) Search(char rune) *Message {
	if char == CharBackspace || char == CharCtrlH {
		if len(o.Filter) > 0 {
			o.Filter = o.Filter[:len(o.Filter)-1]
			return o.index().Render()
		}
	} else if char > 0x1f && char < 0x7f {
		o.Filter += string(char)
		return o.index().Render()
	}
	return o.Cache
}

// index indexes the filtered results (or all options if no filter exists), and
// sets the Hovered index to the first option
// if the currrent Hovered index is not in the filtered list
func (o *Options) index() *Options {
	o.Index = o.Index[:0]
	if len(o.Filter) == 0 {
		for i := 0; i < len(o.Opts); i++ {
			o.Index = append(o.Index, i)
		}
		o.displayLen()
		return o
	}
	hasHovered := false
	for i, opt := range o.Opts {
		if strings.Contains(strings.ToLower(opt), strings.ToLower(o.Filter)) {
			o.Index = append(o.Index, i)
			if i == o.Hovered {
				hasHovered = true
			}
		}
	}
	if len(o.Index) > 0 && !hasHovered {
		o.Hovered = o.Index[0]
	}
	o.displayLen()
	return o
}

// displayLen sets the display length
func (o *Options) displayLen() {
	if l := len(o.Index); l > optDisplayMax {
		o.Display = optDisplayMax
	} else {
		o.Display = l
	}
}

// truncate truncates the options to the given length
func (o *Options) truncate() []int {
	if l := len(o.Index); l > o.Display {
		// get hovered index in index list
		hi := o.HoveredIndex()
		// set left and right indexes
		half := o.Display / 2
		li := hi - half
		ri := hi + half
		switch {
		case li < 0:
			li = 0
			ri = o.Display
		case ri >= l:
			ri = l
			li = ri - o.Display
		default:
			ri = (hi + o.Display) - (hi - li)
		}
		// return the truncated options
		return o.Index[li:ri]
	}
	return o.Index
}
