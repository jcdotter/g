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
	"bytes"
	"os"
	"time"

	"github.com/jcdotter/go/buffer"
)

// --------------------------------------------------------------------------- /
// TODO:
// -[ ] text selection
//   -[ ] store start pos
//   -[ ] provide methods on selection - copy, cut, paste, delete
// --------------------------------------------------------------------------- /

// --------------------------------------------------------------------------- /
// Cursor definitions
// --------------------------------------------------------------------------- /

type Cursor struct {
	c *os.File // terminal where the cursor resides
	d []int    // dimensions of the current prompt
	h pos      // home cursor pos in prompt
	p pos      // current cursor pos in prompt
	s pos      // saved cursor pos in prompt
}

type pos struct {
	x int
	y int
}

// --------------------------------------------------------------------------- /
// Cursor constructors and destructors
// --------------------------------------------------------------------------- /

func NewCursor(terminal *os.File) *Cursor {
	return &Cursor{
		c: terminal,
		d: make([]int, 1, 10),
	}
}

func (c *Cursor) Hide() *Cursor {
	c.c.Write(CursorHide)
	return c
}

func (c *Cursor) Show() *Cursor {
	c.c.Write(CursorShow)
	return c
}

// --------------------------------------------------------------------------- /
// Cursor postioning methods
// --------------------------------------------------------------------------- /

// SetPos unsafely sets the cursor position.
// Will not move cursor, only sets the coordinates.
func (c *Cursor) SetPos(x, y int) *Cursor {
	c.SetRow(y)
	c.SetCol(x)
	return c
}

// SetRow unsafely sets the cursor position.
// Will not move cursor, only sets the y coordinate.
func (c *Cursor) SetRow(y int) *Cursor {
	if lr := len(c.d) - 1; y > lr {
		c.AddRows(y - lr)
	}
	c.p.y = y
	return c
}

// SetCol unsafely sets the cursor position.
// Will not move cursor, only sets the x coordinate.
func (c *Cursor) SetCol(x int) *Cursor {
	if x > c.d[c.p.y] {
		c.AddCols(x - c.d[c.p.y])
	}
	c.p.x = x
	return c
}

// SetToEOS unsafely sets the cursor position to the end
// of the current prompt. Will not move cursor, only
// sets the coordinates.
func (c *Cursor) SetToEOS() *Cursor {
	c.p.y = len(c.d) - 1
	c.p.x = c.d[c.p.y]
	return c
}

func (c *Cursor) Pos() (x, y int) {
	return c.p.x, c.p.y
}

func (c *Cursor) Row() int {
	return c.p.y
}

func (c *Cursor) Col() int {
	return c.p.x
}

func (c *Cursor) SavedPos() (x, y int) {
	return c.s.x, c.s.y
}

func (c *Cursor) SavePos() *Cursor {
	c.s = c.p
	c.c.Write(CursorSave)
	return c
}

func (c *Cursor) RestorePos() *Cursor {
	c.p = c.s
	c.c.Write(CursorRestore)
	return c
}

func (c *Cursor) HomePos() (x, y int) {
	return c.h.y, c.h.y
}

func (c *Cursor) SetHome() *Cursor {
	c.h = c.p
	return c
}

func (c *Cursor) Home() *Cursor {
	if u := c.p.y - c.h.y; u > 0 {
		c.Up(u)
	} else if u < 0 {
		c.Down(-u)
	}
	if l := c.p.x - c.h.x; l > 0 {
		c.Left(l)
	} else if l < 0 {
		c.Right(-l)
	}
	return c
}

// --------------------------------------------------------------------------- /
// Cursor demension methods
// --------------------------------------------------------------------------- /

func (c *Cursor) Append(d []int) *Cursor {
	if len(d) == 0 {
		return c
	}
	r := c.Rows() - 1
	c.d[r] += d[0]
	c.d = append(c.d, d[1:]...)
	return c
}

func (c *Cursor) ScrollDown(n int) *Cursor {
	if n <= 0 {
		return c
	}
	x, y := c.Pos()
	c.EOS()
	_, ey := c.Pos()
	c.c.Write(bytes.Repeat(EOL, n))
	c.SetCol(0)
	time.Sleep(5 * time.Second)
	c.Action(EscUp, n)
	time.Sleep(5 * time.Second)
	c.Up(ey - y).Right(x)
	time.Sleep(5 * time.Second)
	return c
}

func (c *Cursor) AddRows(n int) *Cursor {
	if n <= 0 {
		return c
	}
	c.d = append(c.d, make([]int, n)...)
	return c
}

func (c *Cursor) AddCols(n int) *Cursor {
	if n < 0 {
		return c
	}
	c.d[c.p.y] += n
	return c
}

func (c *Cursor) Rows() int {
	return len(c.d)
}

func (c *Cursor) RowsAbove() int {
	return c.p.y
}

func (c *Cursor) RowsBelow() int {
	return len(c.d) - c.p.y - 1
}

func (c *Cursor) Cols() int {
	return c.d[c.p.y]
}

func (c *Cursor) BlockCols() (cols int) {
	for _, col := range c.d {
		if col > cols {
			cols = col
		}
	}
	return
}

func (c *Cursor) ColsLeft() int {
	return c.p.x
}

func (c *Cursor) ColsRight() int {
	return c.d[c.p.y] - c.p.x
}

// --------------------------------------------------------------------------- /
// Cursor movement methods
// --------------------------------------------------------------------------- /

func (c *Cursor) Action(action byte, n int) *Cursor {
	b := buffer.Pool.Get()
	defer b.Free()
	b.MustWrite(ESC)
	b.MustWriteInt(n)
	b.MustWriteByte(action)
	MustRW(c.c.Write(b.Bytes()))
	return c
}

func (c *Cursor) Up(n int) *Cursor {
	if r := c.RowsAbove(); n > r || n < 0 {
		n = r
	}
	if n == 0 {
		return c
	}
	c.p.y -= n
	if c.p.x > c.d[c.p.y] {
		c.Left(c.p.x - c.d[c.p.y])
	}
	return c.Action(EscUp, n)
}

func (c *Cursor) Down(n int) *Cursor {
	if r := c.RowsBelow(); n < 0 {
		n = r
	} else if d := n - r; d > 0 {
		c.EOS()
		c.c.Write(bytes.Repeat(EOL, d))
		c.d = append(c.d, make([]int, d)...)
		c.SetPos(0, len(c.d)-1)
		return c
	}
	if n == 0 {
		return c
	}
	c.p.y += n
	if c.p.x > c.d[c.p.y] {
		c.Left(c.p.x - c.d[c.p.y])
	}
	return c.Action(EscDown, n)
}

func (c *Cursor) Right(n int) *Cursor {
	if c := c.ColsRight(); n > c || n < 0 {
		n = c
	}
	if n == 0 {
		return c
	}
	c.p.x += n
	return c.Action(EscRight, n)
}

func (c *Cursor) Left(n int) *Cursor {
	if c := c.ColsLeft(); n > c || n < 0 {
		n = c
	}
	if n == 0 {
		return c
	}
	c.p.x -= n
	return c.Action(EscLeft, n)
}

func (c *Cursor) Forward(n int) *Cursor {
	if n < 0 {
		return c.Backward(-n)
	}
	if r := c.ColsRight(); n > r {
		c.Down(1)
		c.Forward(n - r)
		return c
	}
	c.Right(n)
	return c
}

func (c *Cursor) Backward(n int) *Cursor {
	if n < 0 {
		return c.Forward(-n)
	}
	if r := c.ColsLeft(); n > r {
		c.Up(1)
		c.Backward(n - r)
		return c
	}
	c.Left(n)
	return c
}

func (c *Cursor) BOL() *Cursor {
	if cols := c.ColsLeft(); cols > 0 {
		c.Left(cols)
	}
	return c
}

func (c *Cursor) EOL() *Cursor {
	if cols := c.ColsRight(); cols > 0 {
		c.Right(cols)
	}
	return c
}

func (c *Cursor) BOS() *Cursor {
	c.Up(-1)
	c.BOL()
	return c
}

func (c *Cursor) EOS() *Cursor {
	c.Down(-1)
	c.EOL()
	return c
}

// --------------------------------------------------------------------------- /
// Cursor clear methods
// --------------------------------------------------------------------------- /

func (c *Cursor) ClearLine() *Cursor {
	if c.Cols() > 0 {
		c.BOL()
		c.d[c.p.y] = 0
		c.Action(EscClearLn, 2)
	}
	return c
}

func (c *Cursor) Clear() *Cursor {
	if c.Cols() > 0 || c.Rows() > 1 {
		c.BOS()
		c.Action(EscClear, 0)
		c.d = []int{0}
	}
	return c
}

func (c *Cursor) ClearUp(n int) *Cursor {
	if r := c.RowsAbove(); n > r || n < 0 {
		n = r
	}
	for i := 0; i < n; i++ {
		c.Up(1)
		c.ClearLine()
	}
	return c
}

func (c *Cursor) ClearDown(n int) *Cursor {
	if r := c.RowsBelow(); n > r || n < 0 {
		n = r
	}
	for i := 0; i < n; i++ {
		c.Down(1)
		c.ClearLine()
	}
	return c
}

func (c *Cursor) ClearRight() *Cursor {
	if c.ColsRight() > 0 {
		c.d[c.p.y] = c.p.x
		c.Action(EscClearLn, 0)
	}
	return c
}

func (c *Cursor) ClearLeft() *Cursor {
	if r := c.ColsLeft(); r > 0 {
		c.p.x = 0
		c.d[c.p.y] = r
		c.Action(EscClearLn, 1)
	}
	return c
}
