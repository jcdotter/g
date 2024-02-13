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
	"errors"
	"strconv"

	"github.com/jcdotter/go/buffer"
)

var (
	ErrInvalidFlag     = errors.New("invalid flag")
	ErrInvalidFlagType = errors.New("invalid flag type")
)

// --------------------------------------------------------------------------- /
// FlagType definition
// --------------------------------------------------------------------------- /

type FlagType uint8

const (
	BoolFlag FlagType = iota
	IntFlag
	NumFlag
	TextFlag
	CmdFlag
)

var flagTypes = []string{
	"(boolean value)",
	"(integer value)",
	"(numeric value)",
	"(text value)   ",
	"(subcommand)   ",
}

func (ft FlagType) String() string {
	return flagTypes[ft]
}

// --------------------------------------------------------------------------- /
// FlagSet definition
// --------------------------------------------------------------------------- /

type FlagSet struct {
	Cmd   *Command
	Flags []*Flag
	Set   bool
}

func NewFlagSet(c *Command) *FlagSet {
	if c == nil {
		return nil
	}
	return &FlagSet{
		Cmd:   c,
		Flags: make([]*Flag, 0, 8),
	}
}

func (fs *FlagSet) Len() int {
	return len(fs.Flags)
}

func (fs *FlagSet) Reset() {
	for _, f := range fs.Flags {
		if f.Value != nil {
			f.init()
		}
	}
	fs.Set = false
}

func (fs *FlagSet) Clear() {
	fs.Flags = make([]*Flag, 0, 8)
	fs.Set = false
}

func (fs *FlagSet) GetIndex(name string) (*Flag, int) {
	for i, f := range fs.Flags {
		if f.Name == name || f.Short == name {
			return f, i
		}
	}
	for _, p := range fs.Cmd.Parents {
		if f := p.Flags().GetP(name); f != nil {
			return f, -1
		}
	}
	return nil, -1
}

func (fs *FlagSet) Get(name string) *Flag {
	f, _ := fs.GetIndex(name)
	return f
}

func (fs *FlagSet) GetP(name string) *Flag {
	f, _ := fs.GetIndex(name)
	if f != nil && f.Persists {
		return f
	}
	return nil
}

func (fs *FlagSet) IsSet(name string) bool {
	if f := fs.Get(name); f != nil {
		if b, is := f.Value.(*BoolValue); is {
			return bool(*b)
		}
		return true
	}
	return false
}

func (fs *FlagSet) Add(f *Flag) *Flag {
	if f == nil {
		return nil
	}
	if fs == nil {
		*fs = *NewFlagSet(nil)
	}
	Must(f.init())
	if _, i := fs.GetIndex(f.Name); i != -1 {
		fs.Flags[i] = f
		return f
	}
	fs.Flags = append(fs.Flags, f)
	return f
}

func (fs *FlagSet) Remove(name string) {
	if _, i := fs.GetIndex(name); i != -1 {
		fs.Flags = append(fs.Flags[:i], fs.Flags[i+1:]...)
	}
}

func (fs *FlagSet) Parse(args []string) (cmd string, err error) {
	var subs int
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--" || arg == "-":
			fs.Reset()
			return cmd, ErrInvalidFlag
		case arg[:2] == "--":
			i, err = fs.parseLong(args, i)
		case arg[0] == '-':
			i, err = fs.parseShort(args, i)
		case i == 0:
			f := fs.Get(arg)
			if f == nil || f.Type != CmdFlag {
				return arg, nil
			}
			if subs > 0 {
				fs.Reset()
				return cmd, ErrInvalidCommand
			}
			err = f.Value.(*BoolValue).Toggle()
			subs++
		default:
			fs.Reset()
			return cmd, ErrInvalidFlag
		}
		if err != nil {
			return cmd, err
		}
	}
	fs.Set = true
	return
}

func (fs *FlagSet) parseLong(args []string, i int) (int, error) {
	arg := args[i]
	if l := len(arg); arg[:2] == "--" && l > 2 {
		end := 2
		for ; end < l; end++ {
			if arg[end] == '=' || arg[end] == ':' {
				break
			}
		}
		return fs.parse(args, i, 2, end)
	}
	return i, ErrInvalidFlag
}

func (fs *FlagSet) parseShort(args []string, i int) (int, error) {
	arg := args[i]
	if l := len(arg); arg[0] == '-' && l > 1 {
		if l > 2 {
			if sep := arg[2:3]; sep != "=" && sep != ":" {
				return i, ErrInvalidFlag
			}
		}
		return fs.parse(args, i, 1, 2)
	}
	return i, ErrInvalidFlag
}

func (fs *FlagSet) parse(args []string, i, start, end int) (int, error) {
	arg := args[i]
	if f := fs.Get(arg[start:end]); f != nil {
		if astart := end + 1; astart < len(arg) {
			return i, f.SetValue(arg[astart:])
		} else if next := i + 1; next < len(args) && args[next][0] != '-' {
			return i + 1, f.SetValue(args[next])
		} else if f.Type == BoolFlag || f.Type == CmdFlag {
			return i, f.Toggle()
		}
	}
	return i, ErrInvalidFlag
}

func (fs *FlagSet) AddBool(name, short, use string, dflt bool, ptr ...*bool) *Flag {
	return fs.flag(BoolFlag, name, short, use, &dflt, ptr)
}

func (fs *FlagSet) AddInt(name, short, use string, dflt int, ptr ...*int) *Flag {
	return fs.flag(IntFlag, name, short, use, &dflt, ptr)
}

func (fs *FlagSet) AddNum(name, short, use string, dflt float64, ptr ...*float64) *Flag {
	return fs.flag(NumFlag, name, short, use, &dflt, ptr)
}

func (fs *FlagSet) AddText(name, short, use, dflt string, ptr ...*string) *Flag {
	return fs.flag(TextFlag, name, short, use, &dflt, ptr)
}

func (fs *FlagSet) AddCmd(name, short, use string, cmd *Command) *Flag {
	return fs.flag(CmdFlag, name, short, use, cmd, nil)
}

func (fs *FlagSet) flag(typ FlagType, name, short, use string, dflt any, ptr any) *Flag {
	return fs.Add(&Flag{
		Type:     typ,
		Name:     name,
		Short:    short,
		Use:      use,
		Pointers: ptr,
		Default:  dflt,
	})
}

func (fs *FlagSet) String() string {
	l := fs.Len()
	if l == 0 {
		return ""
	}
	b := buffer.Pool.Get()
	defer b.Free()
	for _, flag := range fs.Flags {
		if !flag.Persists {
			b.MustWriteString(flag.String()).MustWrite(EOL)
			l--
		}
	}
	if b.Len() > 0 {
		b.MustPrepend(EOL).MustPrependString("Flags:").MustWrite(EOL)
	}
	if l > 0 {
		b.MustWriteString("Global Flags:").MustWrite(EOL)
		for _, flag := range fs.Flags {
			if flag.Persists {
				b.MustWriteString(flag.String()).MustWrite(EOL)
			}
		}
	}
	return b.String()
}

// --------------------------------------------------------------------------- /
// Flag definition
// --------------------------------------------------------------------------- /

type Flag struct {
	Type     FlagType // The data type of the flag.
	Name     string   // The name of the flag.
	Short    string   // The short name of the flag.
	Use      string   // The usage message of the flag.
	Value    Value    // The value of the flag.
	Pointers any      // The pointers to be populated when the flag is populated.
	Default  any      // The default value of the flag; if a CmdFlag, the Command to be executed.
	Persists bool     // the flag persists to all sub-commands.
}

func (f *Flag) init() error {
	if f.Default == nil {
		return ErrInvalidFlag
	}
	switch f.Type {
	case BoolFlag:
		f.Value = (*BoolValue)(f.Default.(*bool))
	case IntFlag:
		f.Value = (*IntValue)(f.Default.(*int))
	case NumFlag:
		f.Value = (*NumValue)(f.Default.(*float64))
	case TextFlag:
		f.Value = (*Textvalue)(f.Default.(*string))
	case CmdFlag:
		f.Value = new(BoolValue)
	default:
		return ErrInvalidFlagType
	}
	return f.SetPointers()
}

func (f *Flag) Persist() *Flag {
	f.Persists = true
	return f
}

func (f *Flag) String() string {
	b := buffer.Pool.Get()
	defer b.Free()
	b.MustWriteByte('\t')
	if f.Short != "" {
		b.MustWriteByte('-').MustWriteString(f.Short).MustWriteString(", ")
	}
	n := bytes.Repeat([]byte{' '}, 16)
	copy(n, f.Name)
	b.MustWriteString("--").MustWrite(n).MustWriteByte(' ')
	b.MustWriteString(f.Type.String()).MustWriteByte(' ')
	b.MustWriteString(f.Use)
	return b.String()
}

func (f *Flag) GetValue() any {
	if f.Type == CmdFlag {
		if f.Value != nil && *f.Value.(*BoolValue) {
			return f.Default
		}
		return nil
	}
	if f.Value != nil {
		return f.Value
	}
	return f.Default
}

func (f *Flag) SetValue(s string) error {
	if err := f.Value.Set(s); err != nil {
		return err
	}
	return f.SetPointers()
}

func (f *Flag) Toggle() error {
	if b, is := f.Value.(*BoolValue); is {
		if err := b.Toggle(); err != nil {
			return err
		}
		return f.SetPointers()
	}
	return ErrInvalidFlag
}

func (f *Flag) SetPointers() error {
	if f.Pointers != nil {
		switch v := f.Value.(type) {
		case *BoolValue:
			for _, ptr := range f.Pointers.([]*bool) {
				*ptr = bool(*v)
			}
		case *IntValue:
			for _, ptr := range f.Pointers.([]*int) {
				*ptr = int(*v)
			}
		case *NumValue:
			for _, ptr := range f.Pointers.([]*float64) {
				*ptr = float64(*v)
			}
		case *Textvalue:
			for _, ptr := range f.Pointers.([]*string) {
				*ptr = string(*v)
			}
		default:
			return ErrInvalidFlag
		}
	}
	return nil
}

func (f *Flag) Bool() bool {
	if v := f.GetValue(); v != nil {
		if b, is := f.Value.(*BoolValue); is {
			return bool(*b)
		}
		return true
	}
	return false
}

func (f *Flag) Int() int {
	if v := f.GetValue(); v != nil {
		switch v := v.(type) {
		case *BoolValue:
			if bool(*v) {
				return 1
			}
		case *IntValue:
			return int(*v)
		case *NumValue:
			return int(*v)
		case *Textvalue:
			if n, err := strconv.Atoi(string(*v)); err == nil {
				return n
			}
		}
	}
	return 0
}

func (f *Flag) Num() float64 {
	if v := f.GetValue(); v != nil {
		switch v := v.(type) {
		case *BoolValue:
			if bool(*v) {
				return 1
			}
		case *IntValue:
			return float64(*v)
		case *NumValue:
			return float64(*v)
		case *Textvalue:
			if n, err := strconv.ParseFloat(string(*v), 64); err == nil {
				return n
			}
		}
	}
	return 0
}

func (f *Flag) Text() string {
	if v := f.GetValue(); v != nil {
		switch v := v.(type) {
		case *BoolValue:
			if bool(*v) {
				return "true"
			}
		case *IntValue:
			return strconv.Itoa(int(*v))
		case *NumValue:
			return strconv.FormatFloat(float64(*v), 'f', -1, 64)
		case *Textvalue:
			return string(*v)
		}
	}
	return ""
}

func (f *Flag) Cmd() *Command {
	if v := f.GetValue(); v != nil {
		if c, is := v.(*Command); is {
			return c
		}
	}
	return nil
}

// --------------------------------------------------------------------------- /
// Flag Value definition
// --------------------------------------------------------------------------- /

type Value interface {
	String() string
	Set(string) error
}

type BoolValue bool
type IntValue int
type NumValue float64
type Textvalue string

func (b *BoolValue) String() string {
	switch {
	case b == nil:
		return ""
	case bool(*b):
		return "true"
	default:
		return "false"
	}
}

func (i *IntValue) String() string {
	switch {
	case i == nil:
		return ""
	default:
		return strconv.Itoa(int(*i))
	}
}

func (n *NumValue) String() string {
	switch {
	case n == nil:
		return ""
	default:
		return strconv.FormatFloat(float64(*n), 'f', -1, 64)
	}
}

func (t *Textvalue) String() string {
	switch {
	case t == nil:
		return ""
	default:
		return string(*t)
	}
}

func (b *BoolValue) Set(s string) error {
	switch s {
	case "true":
		*b = true
	case "false":
		*b = false
	default:
		return ErrInvalidFlag
	}
	return nil
}

func (i *IntValue) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = IntValue(n)
	return nil
}

func (f *NumValue) Set(s string) error {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = NumValue(n)
	return nil
}

func (t *Textvalue) Set(str string) error {
	if beg, end := str[0], str[len(str)-1]; (beg == '"' && end == '"') || (beg == '\'' && end == '\'') {
		str = str[1 : len(str)-1]
	}
	*t = Textvalue(str)
	return nil
}

func (b *BoolValue) Toggle() error {
	if b == nil {
		*b = true
	}
	*b = BoolValue(!*b)
	return nil
}
