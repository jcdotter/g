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
	"github.com/jcdotter/go/buffer"
)

var ()

// --------------------------------------------------------------------------- /
// Styles
// --------------------------------------------------------------------------- /

// text styles
const (
	Reset Style = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// reset text styles
const (
	ResetBold Style = iota + 22
	ResetItalic
	ResetUnderline
	ResetBlinkSlow
	_
	ResetReverseVideo
	ResetConcealed
	ResetCrossedOut
)

// foreground text colors
const (
	Black Style = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// foreground vibrant text colors
const (
	HiBlack Style = iota + 90
	HiRed
	HiGreen
	HiYellow
	HiBlue
	HiMagenta
	HiCyan
	HiWhite
)

// background text colors
const (
	BGBlack Style = iota + 40
	BGRed
	BGGreen
	BGYellow
	BGBlue
	BGMagenta
	BGCyan
	BGWhite
)

type formatter struct {
	Data any
	Reset, Bold, Faint, Italic, Underline, BlinkSlow, BlinkRapid, ReverseVideo, Concealed, CrossedOut,
	ResetBold, ResetItalic, ResetUnderline, ResetBlinkSlow, ResetReverseVideo, ResetConcealed, ResetCrossedOut,
	Black, Red, Green, Yellow, Blue, Magenta, Cyan, White,
	HiBlack, HiRed, HiGreen, HiYellow, HiBlue, HiMagenta, HiCyan, HiWhite,
	BGBlack, BGRed, BGGreen, BGYellow, BGBlue, BGMagenta, BGCyan, BGWhite []byte
}

func Formatter(data any) *formatter {
	return &formatter{
		data,
		Reset.Code(), Bold.Code(), Faint.Code(), Italic.Code(), Underline.Code(), BlinkSlow.Code(), BlinkRapid.Code(), ReverseVideo.Code(), Concealed.Code(), CrossedOut.Code(),
		ResetBold.Code(), ResetItalic.Code(), ResetUnderline.Code(), ResetBlinkSlow.Code(), ResetReverseVideo.Code(), ResetConcealed.Code(), ResetCrossedOut.Code(),
		Black.Code(), Red.Code(), Green.Code(), Yellow.Code(), Blue.Code(), Magenta.Code(), Cyan.Code(), White.Code(),
		HiBlack.Code(), HiRed.Code(), HiGreen.Code(), HiYellow.Code(), HiBlue.Code(), HiMagenta.Code(), HiCyan.Code(), HiWhite.Code(),
		BGBlack.Code(), BGRed.Code(), BGGreen.Code(), BGYellow.Code(), BGBlue.Code(), BGMagenta.Code(), BGCyan.Code(), BGWhite.Code(),
	}
}

// --------------------------------------------------------------------------- /
// Style type
// --------------------------------------------------------------------------- /

type Style uint8

func (s Style) Bytes() []byte {
	b := buffer.Pool.Get()
	defer b.Free()
	return b.MustWrite(ESC).
		MustWriteUint(uint(s)).
		MustWriteByte(EscStyle).
		Bytes()
}

func (s Style) String() string {
	return string(s.Bytes())
}

func (s Style) Code() []byte {
	return s.Bytes()
}

// --------------------------------------------------------------------------- /
// Styles type
// --------------------------------------------------------------------------- /

type Styles struct {
	buf    *buffer.Buffer
	styles []Style
}

func Styl(styles ...Style) *Styles {
	return (&Styles{
		buf:    buffer.Pool.Get(),
		styles: styles,
	}).setCodes()
}

func (s *Styles) Reset() *Styles {
	s.buf.Reset()
	s.styles = nil
	return s
}

func (s *Styles) Close() {
	s.Reset()
	s.buf.Free()
}

func (s *Styles) Set(styles ...Style) *Styles {
	s.buf.Reset()
	s.styles = styles
	s.setCodes()
	return s
}

func (s *Styles) Add(styles ...Style) *Styles {
	s.styles = append(s.styles, styles...)
	s.setCodes()
	return s
}

func (s *Styles) Remove(styles ...Style) *Styles {
	var removed bool
	for _, remove := range styles {
		for i, style := range s.styles {
			if style == remove {
				s.styles = append(s.styles[:i], s.styles[i+1:]...)
				removed = true
			}
		}
	}
	if removed {
		s.setCodes()
	}
	return s
}

func (s *Styles) setCodes() *Styles {
	s.buf.Reset()
	s.buf.MustWrite(Reset.Code())
	if len(s.styles) > 0 {
		s.buf.MustWrite(ESC)
		for i, style := range s.styles {
			if i > 0 {
				s.buf.MustWriteByte(EscSep)
			}
			s.buf.MustWriteUint(uint(style))
		}
		s.buf.MustWriteByte(EscStyle)
	}
	return s
}

func (s *Styles) Len() int {
	return len(s.styles)
}

func (s *Styles) Bytes() []byte {
	return s.buf.Bytes()
}

func (s *Styles) String() string {
	return s.buf.String()
}

func (s *Styles) Codes() []byte {
	return s.buf.Bytes()
}

func (s *Styles) CodeString() string {
	return s.buf.String()
}

func (s *Styles) Msg(msg ...any) *Message {
	return (&Message{
		buf:    buffer.Pool.Get(),
		dem:    make([]int, 1, 10),
		styles: s,
	}).Add(msg...)
}
