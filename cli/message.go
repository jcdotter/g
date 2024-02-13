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
	"fmt"
	"text/template"

	"github.com/jcdotter/go/buffer"
	"github.com/jcdotter/go/io"
)

// --------------------------------------------------------------------------- /
// Message definition
// --------------------------------------------------------------------------- /

// Message is a constructor for a Cli Message, which
// includes a buffer, template injection, and ansi styles.
type Message struct {
	buf      *buffer.Buffer
	tmp      *template.Template
	styles   *Styles
	dem      []int
	prepared bool
}

// --------------------------------------------------------------------------- /
// Message constructors and destructors
// --------------------------------------------------------------------------- /

// Msg constructs a new Cli Message with the given styles.
func Msg(contents ...any) *Message {
	return (&Message{
		buf:    buffer.Pool.Get(),
		dem:    make([]int, 1, 10),
		styles: Styl(),
	}).Add(contents...)
}

// Template constructs a new Cli Message with the given
// stylized template string and data.
func Tmpl(name, tmpl string, data any) *Message {
	return (&Message{
		buf: buffer.Pool.Get(),
	}).MustParse(name, tmpl).MustExecute(data)
}

// Close returns the Message to the pool.
func (m *Message) Close() {
	m.Reset()
	m.buf.Free()
}

// Reset resets the Message, output and
// attributes to their default values.
func (m *Message) Reset() *Message {
	m.buf.Reset()
	m.styles.Reset()
	m.ResetDem()
	return m
}

// ResetDem resets the Message demensions.
func (m *Message) ResetDem() *Message {
	m.prepared = false
	if len(m.dem) == 0 {
		m.dem = make([]int, 1, 10)
	} else {
		m.dem = m.dem[:1]
		m.dem[0] = 0
	}
	return m
}

// Buffer returns the Message.
func (m *Message) Buffer() *buffer.Buffer {
	return m.buf
}

// --------------------------------------------------------------------------- /
// Message style methods
// --------------------------------------------------------------------------- /

// GetStyle returns the Message styles.
func (m *Message) Style() *Styles {
	return m.styles
}

// Styl sets the Message styles.
func (m *Message) Styl(styles ...Style) *Message {
	m.prepared = false
	return m.SetStyle(styles...)
}

// SetStyle sets the Message styles.
func (m *Message) SetStyle(styles ...Style) *Message {
	m.prepared = false
	m.styles.Set(styles...)
	return m
}

// AddStyle adds the given styles to the Message styles.
func (m *Message) AddStyle(styles ...Style) *Message {
	m.prepared = false
	m.styles.Add(styles...)
	return m
}

// RemoveStyle removes the given styles from the Message styles.
func (m *Message) RemoveStyle(styles ...Style) *Message {
	m.prepared = false
	m.styles.Remove(styles...)
	return m
}

// --------------------------------------------------------------------------- /
// Message write methods
// --------------------------------------------------------------------------- /

// Append appends the given contents to the Message.
func (m *Message) Append(contents ...any) (n int, err error) {
	for i, content := range contents {
		var l int
		if i > 0 {
			if err = m.buf.WriteByte(byte(CharSpace)); err != nil {
				return
			}
			n++
		}
		switch c := content.(type) {
		case byte:
			err = m.AppendByte(c)
			l = 1
		case []byte:
			l, err = m.AppendBytes(c)
		case string:
			l, err = m.AppendString(c)
		case *Message:
			l, err = m.AppendMsg(c)
		case bool:
			l, err = m.AppendBool(c)
		case int:
			l, err = m.AppendInt(c)
		case uint:
			l, err = m.AppendUint(c)
		case float64:
			l, err = m.AppendFloat(c)
		case []string:
			l, err = m.AppendStrings(c...)
		case [][]byte:
			l, err = m.AppendByteSlices(c...)
		default:
			l, err = m.AppendString(fmt.Sprint(c))
		}
		if err != nil {
			return
		}
		n += l
	}
	return
}

// AppendMsg appends the given Message to the Message.
func (m *Message) AppendMsg(msg *Message) (n int, err error) {
	if msg == nil {
		return
	}
	if m.buf.Len() == 0 {
		m.prepared = true
	}
	m.appendDem(msg.dem)
	return m.buf.Write(msg.Bytes())
}

func (m *Message) appendDem(d []int) {
	if len(d) == 0 {
		return
	}
	r := len(m.dem) - 1
	m.dem[r] += d[0]
	m.dem = append(m.dem, d[1:]...)
}

// AppendBytes appends the given bytes to the Message.
func (m *Message) AppendBytes(b []byte) (int, error) {
	m.prepared = false
	return m.buf.Write(b)
}

// AppendByteSlices appends the given byte slices to the Message.
func (m *Message) AppendByteSlices(b ...[]byte) (int, error) {
	m.prepared = false
	return m.buf.WriteByteSlices(b...)
}

// AppendByte appends the given byte to the Message.
func (m *Message) AppendByte(b byte) error {
	return m.buf.WriteByte(b)
}

// AppendBool appends the given bool to the Message.
func (m *Message) AppendBool(b bool) (int, error) {
	return m.buf.WriteBool(b)
}

// AppendInt appends the given int to the Message.
func (m *Message) AppendInt(n int) (int, error) {
	return m.buf.WriteInt(n)
}

// AppendUint appends the given uint to the Message.
func (m *Message) AppendUint(n uint) (int, error) {
	return m.buf.WriteUint(n)
}

// AppendFloat appends the given float to the Message.
func (m *Message) AppendFloat(n float64) (int, error) {
	return m.buf.WriteFloat(n)
}

// AppendString appends the given string to the Message.
func (m *Message) AppendString(s string) (int, error) {
	m.prepared = false
	return m.buf.WriteString(s)
}

// AppendStrings appends the given strings to the Message.
func (m *Message) AppendStrings(s ...string) (int, error) {
	m.prepared = false
	return m.buf.WriteStrings(s...)
}

// Parse parses the given string as a template prepares it for execution.
func (m *Message) Parse(name, s string) (n int, err error) {
	m.tmp, err = template.New(name).Funcs(io.Functions).Parse(s)
	if err != nil {
		return
	}
	return len(s), nil
}

// Execute applies the parsed template to the Message.
func (m *Message) Execute(data any) (n int, err error) {
	if m.tmp == nil {
		return 0, ErrTmplNotFound
	}
	m.prepared = false
	err = m.tmp.Execute(m.buf, Formatter(data))
	if err != nil {
		return
	}
	return m.buf.Len(), nil
}

// --------------------------------------------------------------------------- /
// Must write methods
// --------------------------------------------------------------------------- /

// Append appends the given contents to the Message, and
// panics if there is an error when buffering.
func (m *Message) Add(contents ...any) *Message {
	MustRW(m.Append(contents...))
	return m
}

// Addln appends the given contents and a new line to the Message, and
// panics if there is an error when buffering.
func (m *Message) Addln(contents ...any) *Message {
	m.Add(contents...)
	return m.EOL()
}

// Adds appends any content with style to the Message, and
// panics if there is an error when buffering.
func (m *Message) Adds(content any, s Styles) *Message {
	return m.Add(s.Msg(content))
}

// Addf inserts any content into the format message and
// appends it to the Message. Panics if there is an error when buffering.
func (m *Message) Addf(format string, contents ...any) *Message {
	MustRW(m.AppendString(fmt.Sprintf(format, contents...)))
	return m
}

// Stringf inserts string values into the format message and
// appends it to the Message.
func (m *Message) Stringf(format string, s ...string) *Message {
	MustRW(m.AppendString(Stringf(format, s...)))
	return m
}

// Bool writes the given bool to the Message,
// and panics if there is an error when buffering.
func (m *Message) Bool(b bool) *Message {
	MustRW(m.AppendBool(b))
	return m
}

// Int writes the given int to the Message,
// and panics if there is an error when buffering.
func (m *Message) Int(n int) *Message {
	MustRW(m.AppendInt(n))
	return m
}

// Uint writes the given uint to the Message,
// and panics if there is an error when buffering.
func (m *Message) Uint(n uint) *Message {
	MustRW(m.AppendUint(n))
	return m
}

// Float writes the given float to the Message,
// and panics if there is an error when buffering.
func (m *Message) Float(n float64) *Message {
	MustRW(m.AppendFloat(n))
	return m
}

// Byte writes the given bytes to the Message,
// and panics if there is an error when buffering.
func (m *Message) Byte(b []byte) *Message {
	MustRW(m.AppendBytes(b))
	return m
}

// Text writes the given string to the Message,
// and panics if there is an error when buffering.
func (m *Message) Text(s string) *Message {
	MustRW(m.AppendString(s))
	return m
}

// Space writes a space to the Message,
// and panics if there is an error when buffering.
func (m *Message) Space() *Message {
	Must(m.AppendByte(byte(CharSpace)))
	return m
}

// EOL writes an EOL to the Message,
// and panics if there is an error when buffering.
func (m *Message) EOL() *Message {
	MustRW(m.AppendBytes(EOL))
	return m
}

// EOS writes an EOS to the Message,
// and panics if there is an error when buffering.
func (m *Message) EOS() *Message {
	MustRW(m.AppendBytes(EOS))
	return m
}

// MustParse panics if there is an error when parsing
// the given string as a template.
func (m *Message) MustParse(name, s string) *Message {
	MustRW(m.Parse(name, s))
	return m
}

// MustExecute panics if there is an error when executing
// the parsed template with the given data.
func (m *Message) MustExecute(data any) *Message {
	MustRW(m.Execute(data))
	return m
}

// --------------------------------------------------------------------------- /
// Message read methods
// --------------------------------------------------------------------------- /

// Read reads the Message to the provided byte slice
func (m *Message) Read(b []byte) (n int, err error) {
	return m.Prepare().buf.Read(b)
}

// MustRead reads the Message to provided byte slice
func (m *Message) MustRead(b []byte) *Message {
	MustRW(m.Read(b))
	return m
}

// Bytes returns the Message as a byte slice
func (m *Message) Bytes() []byte {
	return m.Prepare().buf.Bytes()
}

// String returns the Message as a string
func (m *Message) String() string {
	return m.Prepare().buf.String()
}

// Lines returns the number of lines in the Message
func (m *Message) Lines() int {
	return len(m.Prepare().dem)
}

// --------------------------------------------------------------------------- /
// Message prepare for reading methods
// --------------------------------------------------------------------------- /

// PrepFast prepares a message for display in Cli by
// adding the style codes to the message, without cleaning
// the message of non style ANSI escape codes.
func (m *Message) prepFast() *Message {
	MustRW(m.buf.Prepend(m.styles.Codes()))
	MustRW(m.buf.Write(Reset.Code()))
	m.prepared = true
	return m
}

// Prepare prepares a message for display in Cli by
// (i) removing all existing style ANSI escape codes,
// (ii) limiting width of a row to MaxWidth,
// (iii) capturing the dementions of the message, and
// (iv) adding the style codes to the message.
func (m *Message) Prepare() *Message {
	if m.prepared {
		return m
	}
	m.ResetDem()
	if m.buf.Len() == 0 {
		return m
	}
	var cols, col int
	var next bool
	for i := 0; i < m.buf.Len(); i++ {
		cols, i, next = m.prepareLine(i)
		if cols > 0 {
			m.dem[col] += cols
		}
		if next {
			m.dem = append(m.dem, 0)
			col++
		}
	}
	return m.prepFast()
}

// PrepareLine prepares the given bytes for writing to the Cli.
// Removes ANSI escape codes from the given bytes before writing
// them to the Cli. Breaks lines greater than MaxWidth. Returns
// the number of columns in the line, the end index of the line,
// the length of the bytes, and a boolean indicating if there is
// a subsequent line to be prepared.
func (m *Message) prepareLine(start int) (cols, end int, next bool) {
	for end = start; start < m.buf.Len(); start++ {
		switch b := m.buf.Get(end); b {
		case Eol:
			if m.buf.Get(end-1) != Rtn {
				m.buf.InsertByte(end, Rtn)
				return cols, end + 1, true
			}
			return cols, end, true
		case Rtn:
			if post, l := end+1, m.buf.Len(); post == l || m.buf.Get(post) != Eol {
				start, end = removeBytes(m.buf, start, 1)
			}
		case Esc:
			start, end = removeCodes(m.buf, start)
		default:
			if b > 0x1f && b < 0x7f {
				cols++
			}
		}
		if cols > MaxWidth {
			m.buf.Insert(end, EOL)
			return cols, end, true
		}
		end++
	}
	return cols, end, false
}

func removeCodes(from *buffer.Buffer, at int) (start, end int) {
	_, end = EscapeAnsiCode(from.Buffer(), at, from.Len())
	if at := end + 1; at < from.Len() && from.Get(at) == Esc {
		_, end = removeCodes(from, at)
	}
	return removeBytes(from, at, end-at+1)
}

func removeBytes(from *buffer.Buffer, at, n int) (start, end int) {
	from.Delete(at, n)
	at--
	return at, at
}

// RemoveAnsiCodes removes all ANSI escape codes of the given
// types from the given bytes.
func RemoveAnsiCodes(bytes []byte, codeType ...byte) []byte {
	var code byte
	var start int
	var len = len(bytes)
	for end := start; end < len; end++ {
		switch bytes[end] {
		case Esc:
			if code, end = EscapeAnsiCode(bytes, end, len); code == codeType[0] {
				bytes = append(bytes[:start], bytes[end+1:]...)
				start--
				len = len - (end - start)
				end = start
			} else {
				start = end
			}
		}
		start++
	}
	return bytes
}

// EscapeAnsiCode checks if the given bytes contain an ANSI escape code.
// Returns a boolean indicating if the given bytes contain an ANSI escape code,
// and the end index of the code.
func EscapeAnsiCode(bytes []byte, start, len int) (code byte, end int) {
	end = start
	if bytes[start] == Esc {
		end++
		if end < len {
			if bytes[end] == EscCode {
				for end++; end < len; end++ {
					if code = bytes[end]; AnsiEnd(code) {
						return
					}
				}
			}
		}
	}
	return
}

// AnsiEnd checks if the given byte is an ANSI escape code terminator.
func AnsiEnd(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// --------------------------------------------------------------------------- /
// Help methods
// --------------------------------------------------------------------------- /

// Must panics if the given error is not nil.
func MustRW(_ int, err error) {
	if err != nil {
		panic(err)
	}
}

func Stringf(f string, s ...string) string {
	b := buffer.Pool.Get()
	defer b.Free()
	for i := 0; i < len(f); i++ {
		if f[i] == '%' {
			if i+1 < len(f) {
				if f[i+1] == 's' {
					if len(s) > 0 {
						b.WriteString(s[0])
						s = s[1:]
						i++
						continue
					} else {
						panic("invalid string format")
					}
				}
			}
		}
		b.WriteByte(f[i])
	}
	if len(s) > 0 {
		panic("invalid string format")
	}
	return b.String()
}
