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
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jcdotter/go/test"
)

func TestCli(t *testing.T) {
	fmt.Println(Msg(Standard("GRPG")).Styl(HiCyan))
	/* c := New()
	defer c.Close()
	r0, _ := c.Prompt(Input(Msg("Enter Name:")))
	r1, _ := c.Prompt(Select(
		Msg("Select an option:"),
		[]string{
			"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
			"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
			"twenty", "twenty-one"},
	))
	c.Restore()
	fmt.Println(r0, "selected:", r1) */
}

func TestStyle(t *testing.T) {
	gt := test.New(t)
	gt.Msg = "Styles.%s"
	s := Styl(Bold, Underline, HiWhite)
	gt.Equal(3, s.Len(), "Len")
	gt.Equal([]byte("\x1b[0m\x1b[1;4;97m"), s.Bytes(), "Bytes")
	gt.Equal([]byte("\x1b[0m\x1b[1;97m"), s.Remove(Underline).Bytes(), "Remove")
	gt.Equal([]byte("\x1b[0m\x1b[1;97;4m"), s.Add(Underline).Bytes(), "Add")
}

func TestMessage(t *testing.T) {
	gt := test.New(t)
	var exp []byte
	var res *Message

	// Test Message Style
	gt.Msg = "Message.Style.%s"
	exp = []byte("\x1b[0m\x1b[1;97mTEST\x1b[0m")
	res = Styl(Bold, HiWhite).Msg("TEST")
	gt.Equal(string(exp), res.String(), "init")
	exp = []byte("\x1b[0m\x1b[2;90mTEST\x1b[0m")
	res = res.Styl(Faint, HiBlack)
	gt.Equal(string(exp), res.String(), "restyle")

	// Test Message Prepare
	gt.Msg = "Message.Prepare.%s"
	s := []byte("\x1b[0m")
	b := bytes.Repeat([]byte{'A'}, 96)
	exp = append(append(s, append(bytes.Repeat(append(b, []byte("\r\n")...), 9), b...)...), s...)
	res = Msg(bytes.Repeat(append(CursorShow, b...), 10))
	gt.Equal(string(exp), res.String(), "bytes")

	// Test Message Output
	b = []byte("This is \x1b[0m\rtest\x1b[0A\n\x1b[0Atext\n")
	exp = []byte("\x1b[0m\x1b[2;90mThis is test\r\ntext\r\n\x1b[0m")
	res = Styl(Faint, HiBlack).Msg(b)
	gt.Equal(string(exp), res.String(), "bytes")
}

func TestCursor(t *testing.T) {
	gt := test.New(t)
	gt.Msg = "Cursor.%s"

	// Test Cursor Movements
	c := New()
	c.WriteString("Zero\nOne\nTwo\nThree\nFour\nFive\nSix\nSeven\nEight\nNine\nTen\n")
	r0 := c.Cursor().Rows()
	ds := c.Cursor().d
	c.Restore()
	d0 := make([]int, len(ds))
	copy(d0, ds)
	time.Sleep(1 * time.Second)
	c.Cursor().Up(3).Right(3)
	c.Cursor().ClearRight()
	d2 := c.Cursor().d[8]
	x2, y2 := c.Cursor().Pos()
	time.Sleep(1 * time.Second)
	c.Cursor().ClearUp(1)
	r3 := c.Cursor().Rows()
	x3, y3 := c.Cursor().Pos()
	time.Sleep(1 * time.Second)
	c.Cursor().Clear()
	r4 := c.Cursor().Rows()
	x4, y4 := c.Cursor().Pos()
	c.Close()

	// Validate Cursor Movements
	gt.Equal(12, r0, "Write.Rows")
	gt.Equal([]int{4, 3, 3, 5, 4, 4, 3, 5, 5, 4, 3, 0}, d0, "Write.d")
	gt.Equal(3, d2, "ClearRight.d")
	gt.Equal(3, x2, "ClearRight.Pos.x")
	gt.Equal(8, y2, "ClearRight.Pos.y")
	gt.Equal(12, r3, "ClearUp.Rows")
	gt.Equal(0, x3, "ClearUp.Pos.x")
	gt.Equal(7, y3, "ClearUp.Pos.y")
	gt.Equal(1, r4, "Clear.Rows")
	gt.Equal(0, x4, "Clear.Pos.x")
	gt.Equal(0, y4, "Clear.Pos.y")

}

func TestWrite(t *testing.T) {
	gt := test.New(t)

	gt.Msg = "Message.Write.%s"
	exp := []byte("\x1b[0mTest message...\x1b[0m")
	c := New()
	defer c.Close()
	m := Msg("Test message...")
	n, err := c.WriteMsg(m)
	gt.Equal(exp, m.Bytes(), "bytes")
	gt.Equal(nil, err, "err")
	gt.Equal(len(exp), n, "len")
	gt.Equal([]int{15}, c.cur.d, "dem")
	gt.Equal(15, c.cur.p.x, "cur.x")
	gt.Equal(0, c.cur.p.y, "cur.y")

	gt.Msg = "Message.Write.Instructions.%s"
	c.cur.SetHome()
	c.cur.Down(2)
	c.WriteMsg(instInput)
	c.cur.Home()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	time.Sleep(500 * time.Millisecond)
	go func() {
		time.Sleep(time.Millisecond)
		c.Input([]byte{Rtn}, 300)
		wg.Done()
	}()
	gt.Equal([]int{15, 0, 27}, c.cur.d, "dem")
	gt.Equal(15, c.cur.p.x, "cur.x")
	gt.Equal(0, c.cur.p.y, "cur.y")
	c.ReadLine()
	wg.Wait()
	c.Restore()

	gt.Msg = "Message.Write.Clear.%s"
	c.cur.Clear()
	gt.Equal([]int{0}, c.cur.d, "dem")
	gt.Equal(0, c.cur.p.x, "cur.x")
	gt.Equal(0, c.cur.p.y, "cur.y")
}

func TestOptions(t *testing.T) {
	gt := test.New(t)
	opts := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	opt := NewOptions(opts, false)
	opt.Hovered = 5

	gt.Msg = "Options.Truncate - %s"
	gt.Equal([]int{2, 3, 4, 5, 6, 7, 8}, opt.truncate(), "cursor mid")
	opt.Hovered = 0
	gt.Equal([]int{0, 1, 2, 3, 4, 5, 6}, opt.truncate(), "cursor left")
	opt.Hovered = 9
	gt.Equal([]int{3, 4, 5, 6, 7, 8, 9}, opt.truncate(), "cursor right")

	gt.Msg = "Options.Filter.%s"
	opt.Search('e')
	gt.Equal([]int{0, 2, 4, 6, 7, 8, 9}, opt.Index, "Index")
	gt.Equal(9, opt.Hovered, "Hovered")

	gt.Msg = "Options.Render"
	options := "" +
		"\r\n\x1b[0m\x1b[3;90mshowing search results for: 'e'\r\n\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mone\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mthree\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mfive\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mseven\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37meight\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mnine\r\n" +
		"\x1b[0m\x1b[1;97m> \x1b[0m\x1b[1;97mten\r\n" +
		"\x1b[0m"
	gt.Equal(options, opt.Render().String())

	gt.Msg = "Options.Up"
	opt.Up()
	options = "" +
		"\r\n\x1b[0m\x1b[3;90mshowing search results for: 'e'\r\n\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mone\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mthree\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mfive\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mseven\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37meight\r\n" +
		"\x1b[0m\x1b[1;97m> \x1b[0m\x1b[1;97mnine\r\n" +
		"\x1b[0m\x1b[2;37m  \x1b[0m\x1b[2;37mten\r\n" +
		"\x1b[0m"
	gt.Equal(options, opt.Render().String())

}

func TestPrompt(t *testing.T) {
	gt := test.New(t)
	gt.Msg = "Prompt.%s"
	var res any
	var err error

	// Test Input
	res, err = MimicInput("Testing")
	gt.Equal(nil, err, "Error")
	gt.Equal("Testing", res, "Input")

	// Test Password
	res, err = MimicPassword("myp@55w0rd")
	gt.Equal(nil, err, "Error")
	gt.Equal("myp@55w0rd", res, "Password")

	// Test Select
	options := []string{
		"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
		"twenty", "twenty-one"}
	inputs := []byte{'e', 'e', byte(simDownArrow), byte(simDownArrow), byte(CharReturn)}
	res, err = MimicSelect(options, inputs)
	gt.Equal(nil, err, "Error")
	gt.Equal("fourteen", res, "Select")

	// Test MultiSelect
	inputs = []byte{'e', 'e', byte(simDownArrow), byte(CharSpace), byte(simDownArrow), byte(simDownArrow), byte(CharSpace), byte(CharReturn)}
	res, err = MimicMultiSelect(options, inputs)
	gt.Equal(nil, err, "Error")
	gt.Equal([]string{"thirteen", "fifteen"}, res, "MultiSelect")

}

func TestFlag(t *testing.T) {
	gt := test.New(t)

	// Test Flag
	gt.Msg = "Flag.%s"
	f := &Flag{
		Type:    BoolFlag,
		Name:    "test",
		Short:   "t",
		Use:     "test flag",
		Default: new(bool),
	}
	gt.Equal(nil, f.init(), "initValue")
	gt.NotEqual(nil, f.Value, "Value not nil")
	gt.Equal(nil, f.Value.(*BoolValue).Toggle(), "Value.Toggle")
	gt.True(f.Bool(), "Value toggled")

	// Test FlagSet
	gt.Msg = "FlagSet.%s"
	c := &Command{Name: "test"}
	fs := NewFlagSet(c)
	c.flags = fs
	fs.AddText("test", "t", "test flag", "test flag").Persist()

	gt.NotEqual((*Command)(nil), fs.Cmd, "Cmd not nil")
	gt.NotEqual(nil, fs.Flags, "Flags not nil")
	gt.False(fs.Set, "Set (not)")
	gt.Equal(1, fs.Len(), "Len()")
	gt.Equal("test", fs.Get("test").Name, "Name")
	gt.Equal(TextFlag, fs.Get("test").Type, "Type")
	gt.True(fs.Get("test").Persists, "Persists")

	// Test FlagSet default values
	gt.Msg = "FlagSet.Default.%s"
	cmd, err := fs.Parse([]string{})
	gt.Equal(nil, err, "Error")
	gt.Equal("", cmd, "Command")
	gt.True(fs.Set, "Set")
	gt.Equal("test flag", fs.Get("test").Text(), "Value")

	// Test Parse command args to FlagSet
	gt.Msg = "FlagSet.Parse.%s"
	cmd, err = fs.Parse([]string{"-t", "true"})
	gt.Equal(nil, err, "Error")
	gt.Equal("", cmd, "Command")
	gt.True(fs.Set, "Set")
	gt.Equal("true", fs.Get("test").Text(), "Value")

	// Test Persistent Flags
	c1 := &Command{Name: "test1", Parents: []*Command{c}}
	c2 := &Command{Name: "test2", Parents: []*Command{c1}}
	fs2 := NewFlagSet(c2)

	gt.Msg = "FlagSet.Persist.%s"
	f, i := fs2.GetIndex("test")
	gt.Equal(-1, i, "Get.Index")
	gt.NotEqual((*Flag)(nil), f, "Get.Flag")
	gt.Equal("test", f.Name, "Get.Name")

	cmd, err = fs2.Parse([]string{"--test", "tested"})
	gt.Equal(nil, err, "Error")
	gt.Equal("", cmd, "Command")
	gt.True(fs2.Set, "Set")
	gt.Equal("tested", fs2.Get("test").Text(), "Value")

}

func TestCommand(t *testing.T) {
	gt := test.New(t)

	// Setup test
	name := "test"
	Stdout, _ = os.OpenFile("test.txt", os.O_CREATE|os.O_RDWR, 0755)
	buf := make([]byte, 1024)
	c := &Command{Name: name}
	err := c.execute([]string{name, "--version"})

	// Test Defaults
	gt.Msg = "Command.%s"
	gt.Equal(nil, err, "Execute.Error")
	gt.Equal(name, c.Name, "Name")
	gt.Equal(name+use, c.Use.buf.String(), "Usage")
	gt.Equal(name+description, c.Description.buf.String(), "Description")
	gt.Equal("\x1b[0mv0.0.0\x1b[0m", c.Version.buf.String(), "Version")

	// Test Version Command
	gt.Msg = "Command.Execute.%s"
	Stdout.Seek(0, 0)
	i, err := Stdout.Read(buf)
	gt.Equal(nil, err, "Terminal.Error")
	gt.NotEqual(0, i, "BytesWritten")
	gt.Equal("test \x1b[0mv0.0.0\x1b[0m\r\n", string(buf[:i]), "Version")

	// Test Help Command
	c.Flags().Reset()
	Stdout.Seek(0, 0)
	exp := "\x1b[0mtest is a custom command line application.\x1b[0m\r\n\r\nUsage:\r\n\t\x1b[0mtest <command>... [-a | --arg | --arg=value | --arg:value | -arg value...]\x1b[0m\r\n\r\nGlobal Flags:\r\n\t-v, --version          (subcommand)    display the command version\r\n\t-h, --help             (subcommand)    display the command help\r\n"
	err = c.execute([]string{name, "--help"})
	gt.Equal(nil, err, "Help.Error")
	Stdout.Seek(0, 0)
	i, err = Stdout.Read(buf)
	gt.Equal(nil, err, "Terminal.Error")
	gt.NotEqual(0, i, "BytesWritten")
	gt.Equal(exp, string(buf[:i]), "Help")
	Stdout.Close()
	os.Remove("test.txt")
}

func MimicInput(input string) (res any, err error) {
	c := New()
	defer c.Close()
	return mimic(c, Input(Msg("Enter input: ")), []byte(input), 300)
}

func MimicPassword(password string) (res any, err error) {
	c := New()
	defer c.Close()
	return mimic(c, Password(Msg("Enter password: ")), []byte(password), 50)
}

func MimicSelect(options []string, input []byte) (res any, err error) {
	c := New()
	defer c.Close()
	return mimic(c, Select(Msg("Select option:"), options), []byte(input), 900)
}

func MimicMultiSelect(options []string, input []byte) (res any, err error) {
	c := New()
	defer c.Close()
	return mimic(c, MultiSelect(Msg("Select options:"), options), []byte(input), 900)
}

func mimic(c *Cli, prompt *Prompt, input []byte, dur int) (res any, err error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		time.Sleep(time.Millisecond)
		c.Input(input, dur)
		wg.Done()
	}()
	res, err = c.Prompt(prompt)
	wg.Wait()
	return
}

func TestKeyValue(t *testing.T) {
	c := New()
	defer c.Close()
	c.Write([]byte("press key: "))
	c.cur.SavePos()
	printKeystroke(c)
	c.cur.Clear()
}

func printKeystroke(c *Cli) {
	r, _, _ := c.ReadRune()
	if r == CharCtrlJ {
		return
	}
	c.cur.Hide()
	c.cur.RestorePos()
	c.cur.ClearRight()
	c.Write([]byte(fmt.Sprint("rune: ", r, "; bytes: ", []byte(string(r)))))
	printKeystroke(c)
}

func TestFont(t *testing.T) {
	fmt.Println(string(Graffiti("!-/9 AZa.~")))
}

func TestGenChars(t *testing.T) {
	chars := graffiti.Chars
	l := 0
	for i, b := range chars[0] {
		if b == 0x20 && charEnd(chars, i) {
			fmt.Print(",", i)
			l++
		}
	}
	fmt.Println()
	fmt.Println("len:", l)
}

func charEnd(chars [][]byte, col int) bool {
	for _, r := range chars {
		if r[col] != 0x20 {
			return false
		}
	}
	return true
}

func TestPrintHelp(t *testing.T) {
	c := &Command{
		Name:   "CreateApp",
		Banner: Msg(Graffiti("CreateApp")).Styl(HiCyan),
		Description: Styl(Italic, Faint, HiWhite).Msg(`CreateApp is a test command line application.
		
This is an example of a mutiple-line, long description of an application. The TestApp is
built using various cli libraries, including styles, prompts and custom ASCII art.`),
		Version: Msg("v0.1.2").Styl(Faint, Underline, Green),
	}
	c.Flags().AddText("name", "n", "the name of the app being created", "myapp")
	c.Flags().AddText("author", "a", "the name of the original app author", "decoder")
	c.AddCommand(&Command{
		Name:        "options",
		Description: Msg("options displays the available options"),
		Run: func(cmd *Command, args *FlagSet) error {
			return nil
		},
	})
	c.AddCommand(&Command{
		Name:        "guide",
		Description: Msg("guide guides the user through the app creation process"),
		Run: func(cmd *Command, args *FlagSet) error {
			return nil
		},
	})
	c.execute([]string{"TestApp", "--help"})
}
