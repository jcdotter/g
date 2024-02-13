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
	"errors"
	"os"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/term"
)

// --------------------------------------------------------------------------- /
// Cli variables
// --------------------------------------------------------------------------- /

// error code
var (
	ErrEOF          error = errors.New("EOF")
	ErrEscape       error = errors.New("escape")
	ErrTmplNotFound error = errors.New("template not found")
	ErrRange        error = errors.New("index out of range")
)

// terminal file
var (
	Stdin  *os.File = os.Stdin
	Stdout *os.File = os.Stdout
)

// max terminal width
const MaxWidth int = 96

// special character
const (
	CharCtrlH      rune = 8   // Backspace '\b'
	CharTab        rune = 9   // Tab '\t'
	CharCtrlJ      rune = 10  // NewLine '\n'
	CharReturn     rune = 13  // CarriageReturn '\r'
	CharEsc        rune = 27  // Escape '\x1b'
	CharSpace      rune = 32  // Space ' '
	CharBackspace  rune = 127 // Delete '\x7f'
	simDownArrow   rune = 143
	simUpArrow     rune = 144
	CharCtrlDown   rune = 16949
	CharCtrlUp     rune = 16693
	CharCtrlLeft   rune = 17461
	CharCtrlRight  rune = 17205
	CharCtrlDelete rune = 32309
	CharDownArrow  rune = 4348699
	CharUpArrow    rune = 4283163
	CharLeftArrow  rune = 4479771
	CharRightArrow rune = 4414235
	CharEnd        rune = 4610843
	CharHome       rune = 4741915
	CharPageUp     rune = 2117425947
	CharPageDown   rune = 2117491483
	CharDelete     rune = 2117294875 // delete key (not backspace)
)

// ansi escape code
var (
	ESC           []byte = []byte("\x1b[")
	ESCX          []byte = []byte("\033[")
	CursorShow    []byte = []byte("\x1b[?25h")
	CursorHide    []byte = []byte("\x1b[?25l")
	CursorSave    []byte = []byte("\x1b[s")
	CursorRestore []byte = []byte("\x1b[u")
	Return        []byte = []byte("\r")
	EOL           []byte = []byte("\r\n")
	EOS           []byte = []byte("\r\n\r\n")
)

// ansi escape code
const (
	Eol        byte = '\n'
	Rtn        byte = '\r'
	Esc        byte = '\x1b'
	EscCode    byte = '['
	EscUp      byte = 'A'
	EscDown    byte = 'B'
	EscRight   byte = 'C'
	EscLeft    byte = 'D'
	EscClear   byte = 'J'
	EscClearLn byte = 'K'
	EscStyle   byte = 'm'
	EscSep     byte = ';'
)

// --------------------------------------------------------------------------- /
// Cli definition
// --------------------------------------------------------------------------- /

// Cli is a buffered command line interface that supports
// ANSI escape codes, color and style attributes
// prompt and input methods, and more,
// with buffering for efficient output.
type Cli struct {
	src   *os.File
	fd    int
	term  *term.Terminal
	state *term.State // base terminal state
	cur   *Cursor
	resp  string // last user prompt response
}

// --------------------------------------------------------------------------- /
// Cli getters and setters
// --------------------------------------------------------------------------- /

// New constructs a new Cli with an initial Message.
// If no terminal is provided, os.Stdin is used.
func New(terminal ...*os.File) *Cli {
	use := Stdout
	if len(terminal) > 0 {
		use = terminal[0]
	}
	fd := int(use.Fd())
	state, err := term.MakeRaw(fd)
	Must(err)
	return &Cli{
		src:   use,
		fd:    fd,
		state: state,
		term:  term.NewTerminal(use, ""),
		cur:   NewCursor(use),
	}
}

// Restore restores the Cli's terminal state.
func (c *Cli) Restore() error {
	return term.Restore(c.fd, c.state)
}

// MustRestore restores the Cli's terminal state.
// Panics if the terminal state cannot be restored.
func (c *Cli) MustRestore() {
	Must(c.Restore())
}

// ClearStyle clears the Cli's current style.
func (c *Cli) ClearStyle() {
	MustRW(c.src.Write(Reset.Bytes()))
	c.cur.Show()
}

// Reset resets the Cli's terminal state.
func (c *Cli) Reset() {
	c.cur.Clear()
	c.ClearStyle()
	c.MustRestore()
}

// Close closes cli and returns and error if
// the terminal state cannot be restored.
func (c *Cli) Close(msg ...*Message) error {
	if err := c.Restore(); err != nil {
		return err
	}
	if len(msg) > 0 {
		c.WriteMsg(msg[0])
	}
	c.ClearStyle()
	c.src = nil
	c.fd = 0
	c.term = nil
	c.state = nil
	c.resp = ""
	return nil
}

// Cursor returns the Cli's cursor.
func (c *Cli) Cursor() *Cursor {
	return c.cur
}

// --------------------------------------------------------------------------- /
// Cli write methods
// --------------------------------------------------------------------------- /

// WriteMsg outputs the given Message to the Cli.
func (c *Cli) WriteMsg(m *Message) (n int, err error) {
	if m != nil {
		m.Prepare()
		c.cur.EOS().Append(m.dem)
		defer c.cur.SetToEOS()
		return c.src.Write(m.buf.Buffer())
	}
	return
}

// Write writes the given bytes to the Cli.
// Removes ANSI escape codes from the given bytes
// before writing them to the Cli.
func (c *Cli) Write(p []byte) (n int, err error) {
	m := Msg()
	if _, err = m.AppendBytes(p); err == nil {
		return c.WriteMsg(m)
	}
	return
}

// WriteString writes the given string to the Cli.
func (c *Cli) WriteString(s string) (n int, err error) {
	return c.Write([]byte(s))
}

// Key mimics a user keystroke of char in a cli.
func (c *Cli) Key(char byte) (err error) {
	syscall.RawSyscall(
		syscall.SYS_IOCTL,
		c.src.Fd(),
		syscall.TIOCSTI,
		uintptr(unsafe.Pointer(&char)),
	)
	return
}

// Input mimics a user input of input in a cli with the
// given duration in milliseconds between keystrokes.
func (c *Cli) Input(input []byte, dur int) (err error) {
	t := time.Duration(dur) * time.Millisecond
	var eol bool
	for _, b := range input {
		time.Sleep(t)
		if err = c.Key(b); err != nil {
			return
		}
		if b == Eol || b == Rtn {
			eol = true
			break
		}
	}
	if !eol {
		time.Sleep(t)
		c.Key(Rtn)
	}
	return
}

// --------------------------------------------------------------------------- /
// Cli read methods
// --------------------------------------------------------------------------- /

// ReadRune reads a single keystroke from the Cli as a rune.
func (c *Cli) ReadRune() (rune, int, error) {
	b := make([]byte, 4)
	_, err := c.src.Read(b)
	if err != nil {
		return 0, 0, err
	}
	return *(*rune)(unsafe.Pointer(&b[0])), len(b), nil
}

// ReadLine reads a line of input from the Cli.
func (c *Cli) ReadLine() (res string, err error) {
	c.cur.Show()
	defer c.cur.SetPos(0, c.cur.Row()+1)
	return c.term.ReadLine()
}

// ReadPassword reads a password from the Cli.
func (c *Cli) ReadPassword(s string) (string, error) {
	c.cur.Hide()
	defer c.cur.SetPos(0, c.cur.Row()+1)
	return c.term.ReadPassword(s)
}

// --------------------------------------------------------------------------- /
// Cli prompt methods
// --------------------------------------------------------------------------- /

func (c *Cli) Prompt(p *Prompt) (res any, err error) {
	res, err = p.Prompt(c)
	if err != nil {
		return "", err
	}
	return
}

// --------------------------------------------------------------------------- /
// Cli help methods
// --------------------------------------------------------------------------- /

// Must panics if the given error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
