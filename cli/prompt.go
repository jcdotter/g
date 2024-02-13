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

// --------------------------------------------------------------------------- /
// Prompt types
// --------------------------------------------------------------------------- /

type PromptType uint8

const (
	InputType PromptType = iota
	PasswordType
	MultilineType
	SelectType
	MultiSelectType
)

// --------------------------------------------------------------------------- /
// Prompt functions
// --------------------------------------------------------------------------- /

type PromptFunc func(*Cli) (any, error)

type Callback func(cli *Cli, msg *Message, res any) (any, error)

func NewCallback(fn ...Callback) *Callback {
	var f *Callback
	if len(fn) > 0 {
		f = &fn[0]
	}
	return f
}

// --------------------------------------------------------------------------- /
// Prompt definition
// --------------------------------------------------------------------------- /

// prompt instructions
var (
	instStyl   = Styl(Italic, HiBlack)
	instInput  = instStyl.Msg("press enter ↵ to continue...")
	instSelect = instStyl.Msg("use the arrow keys ↑ ↓ to navigate;\r\npress enter ↵ to select")
	instMulti  = instStyl.Msg("use the arrow keys ↑ ↓ to navigate;\r\npress space bar to select;\r\npress enter ↵ to continue")
)

type Prompt struct {
	cli     *Cli
	Type    PromptType // the kind of prompt
	Msg     *Message   // the prompt message
	Inst    *Message   // the instructions message
	Prompt  PromptFunc // the action to perform when prompting
	Call    *Callback  // user defined func called on prompt input
	Options *Options   // the options for a select prompt
}

// --------------------------------------------------------------------------- /
// Prompt constructors and destructors
// --------------------------------------------------------------------------- /

// Input returns a new input prompt with the given message and callback,
// if any, ready for use as a param in Cli.Prompt(). The Input prompt
// returns the user input as a string.
func Input(msg *Message, fn ...Callback) *Prompt {
	return (&Prompt{
		Type: InputType,
		Msg:  msg,
		Inst: instInput,
		Call: NewCallback(fn...),
	}).Init()
}

// Password returns a new masked prompt with the given message and callback,
// if any, ready for use as a param in Cli.Prompt(). The Password prompt
// returns the user masked input as a string.
func Password(msg *Message, fn ...Callback) *Prompt {
	return (&Prompt{
		Type: PasswordType,
		Msg:  msg,
		Inst: instInput,
		Call: NewCallback(fn...),
	}).Init()
}

// Select returns a new select menu prompt with the given message,
// select options and callback, if any, ready for use as a param
// in Cli.Prompt(). The select prompt returns the user selected
// value as a string.
func Select(msg *Message, opts []string, fn ...Callback) *Prompt {
	return (&Prompt{
		Type:    SelectType,
		Msg:     msg,
		Inst:    instSelect,
		Call:    NewCallback(fn...),
		Options: NewOptions(opts, false),
	}).Init()
}

// MultiSelect returns a new multi-select menu prompt with the given message,
// select options and callback, if any, ready for use as a param in Cli.Prompt().
// The multi-select prompt returns the user selected values as a []string.
func MultiSelect(msg *Message, opts []string, fn ...Callback) *Prompt {
	return (&Prompt{
		Type:    MultiSelectType,
		Msg:     msg,
		Inst:    instMulti,
		Call:    NewCallback(fn...),
		Options: NewOptions(opts, true),
	}).Init()
}

func (p *Prompt) Init() *Prompt {
	p.SetPrompt()
	return p
}

func (p *Prompt) SetPrompt() *Prompt {
	var fn PromptFunc
	switch p.Type {
	case InputType:
		fn = p.Input
	case PasswordType:
		fn = p.Password
	case SelectType, MultiSelectType:
		fn = p.Select
	}
	p.Prompt = fn
	return p
}

func (p *Prompt) Close() {
	p.Clear()
	p.cli.cur.Show()
	p.Msg.Close()
}

func (p *Prompt) Clear() *Prompt {
	p.cli.cur.Clear()
	return p
}

// --------------------------------------------------------------------------- /
// Prompt methods
// --------------------------------------------------------------------------- /

func (p *Prompt) Input(c *Cli) (res any, err error) {
	defer p.Close()
	p.cli = c
	if _, err = p.WritePrompt(); err != nil {
		return "", err
	}
	if res, err = c.ReadLine(); err != nil {
		return "", err
	}
	if p.Call != nil {
		return (*p.Call)(c, p.Msg, res)
	}
	return
}

func (p *Prompt) Password(c *Cli) (res any, err error) {
	defer p.Close()
	p.cli = c
	if _, err = p.WritePrompt(); err != nil {
		return "", err
	}
	if res, err = c.ReadPassword(""); err != nil {
		return "", err
	}
	if p.Call != nil {
		return (*p.Call)(c, p.Msg, res)
	}
	return
}

func (p *Prompt) Select(c *Cli) (res any, err error) {
	defer p.Close()
	p.cli = c
	c.cur.Hide()
	if res, err = p.selectIter(c); err != nil {
		return "", err
	}
	if p.Type == SelectType {
		if res == nil {
			res = ""
		} else {
			res = res.([]string)[0]
		}
	}
	if p.Call != nil {
		return (*p.Call)(c, p.Msg, res)
	}
	return
}

func (p *Prompt) selectIter(c *Cli) (res []string, err error) {
	p.Clear()
	if _, err = p.WritePrompt(); err != nil {
		return nil, err
	}
	var key rune
	if key, _, err = c.ReadRune(); err != nil {
		return nil, err
	}
	c.cur.SetCol(c.cur.Col() + 1)
	switch key {
	case CharCtrlJ, CharReturn:
		return p.Options.Select(), nil
	case CharDownArrow, simDownArrow:
		p.Options.Down()
	case CharUpArrow, simUpArrow:
		p.Options.Up()
	case CharCtrlDown, CharEnd:
		p.Options.Last()
	case CharCtrlUp, CharHome:
		p.Options.First()
	case CharPageDown:
		p.Options.PageDown()
	case CharPageUp:
		p.Options.PageUp()
	case CharSpace:
		p.Options.Toggle()
	default:
		p.Options.Search(key)
	}
	return p.selectIter(c)
}

func (p *Prompt) WritePrompt() (n int, err error) {
	if n, err = p.cli.WriteMsg(p.Msg); err != nil {
		return
	}
	p.cli.cur.SetHome()
	if p.Options != nil {
		if n, err = p.cli.WriteMsg(p.Options.Render()); err != nil {
			return
		}
	}
	p.cli.cur.Down(2)
	p.cli.WriteMsg(p.Inst)
	p.cli.cur.Home()
	return
}
