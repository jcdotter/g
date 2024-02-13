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
	"os"

	"github.com/jcdotter/go/buffer"
)

// errors
var (
	ErrInvalidCommand = errors.New("invalid command")
)

// --------------------------------------------------------------------------- /
// Command default templates
// --------------------------------------------------------------------------- /

var (
	description = " is a custom command line application."
	use         = " <command>... [-a | --arg | --arg=value | --arg:value | -arg value...]"

	// version defaults
	version     = "v0.0.0"
	versionFlag = &Flag{
		Type:     CmdFlag,
		Name:     "version",
		Short:    "v",
		Use:      "display the command version",
		Value:    new(BoolValue),
		Default:  versionCmd,
		Persists: true,
	}
	versionCmd = &Command{
		Name: "version",
		Use:  Msg("display the command version"),
		Run: func(cmd *Command, args *FlagSet) error {
			b := buffer.Pool.Get()
			defer b.Free()
			b.MustWriteString(cmd.Name).MustWriteByte(' ').
				MustWrite(cmd.Version.Bytes()).MustWrite(EOL)
			_, err := Stdout.Write(b.Bytes())
			return err
		},
	}

	// help defaults
	helpFlag = &Flag{
		Type:     CmdFlag,
		Name:     "help",
		Short:    "h",
		Use:      "display the command help",
		Value:    new(BoolValue),
		Default:  helpCmd,
		Persists: true,
	}
	helpCmd = &Command{
		Name: "help",
		Use:  Msg("display the command help"),
		Run: func(cmd *Command, args *FlagSet) error {
			Stdout.Write(cmd.Help.MustExecute(cmd).buf.Bytes())
			return nil
		},
	}
	helpTmpl = `
{{- if .Data.Banner }}{{ .Data.Banner }}{{ "\r\n" }}{{ end }}
{{- .Data.Description }}{{ "\r\n" }}{{ "\r\n" -}}
Usage:{{ "\r\n\t" }}{{.Data.Use}}{{ "\r\n\r\n" }}
{{- if .Data.Example }}Example:{{ "\r\n\t" }}{{ .Data.Example }}{{ "\r\n" }}{{ end }}
{{- if .Data.Subs }}{{ .Data.Subs }}{{ "\r\n" }}{{ end }}
{{- if .Data.Flags }}{{ .Data.Flags }}{{ end }}`
)

// --------------------------------------------------------------------------- /
// Command definition
// --------------------------------------------------------------------------- /

type CmdFunc func(cmd *Command, args *FlagSet) error

// Command is a command line command.
type Command struct {
	init        bool     // the command has been initialized
	Name        string   // the string name of the command
	Version     *Message // the version of the command
	Banner      *Message // the banner of the command
	Short       *Message // the short description of the command
	Description *Message // the description of the command
	Use         *Message // the usage description of the command
	Help        *Message // the help is a template for generating help
	Example     *Message // the example description of the command
	Run         CmdFunc  // the function executed when the command is run
	flags       *FlagSet // the available argurments for the command
	Parents     Commands // the parent commands of the command
	Subs        Commands // the sub commands of the command
	Persist     bool     // the command is executed before its sub commands are executed
}

func (c *Command) initDefaults() {
	if c.Name == "" {
		c.Name = os.Args[0]
	}

	// set the default description
	if c.Description == nil {
		c.Description = Msg(c.Name + description)
	}

	// set the shorr description to the first line
	// of the description, not to exceed 64 bytes
	if c.Short == nil {
		i, d := 0, c.Description.buf.Bytes()
		for ; i < 64 && i < len(d); i++ {
			if d[i] == '\n' || d[i] == '\r' {
				break
			}
		}
		c.Short = Msg(string(d[:i])).Styl(c.Description.styles.styles...)
	}
	if c.Use == nil {
		c.Use = Msg(c.Name + use)
	}

	// set the version flag and command
	if c.Version == nil {
		c.Version = Msg(version)
	}
	if f := c.Flags().Get("version"); f == nil {
		c.Flags().Flags = append(c.Flags().Flags, versionFlag)
	}

	// set the help flag and command
	if c.Help == nil {
		c.Help = Msg().MustParse("help", helpTmpl)
	}
	if f := c.flags.Get("help"); f == nil {
		c.Flags().Flags = append(c.Flags().Flags, helpFlag)
	}
	c.init = true
}

func (c *Command) Execute() error {
	if c == nil {
		return nil
	}
	return c.execute(os.Args)
}

func (c *Command) execute(args []string) error {

	// make sure required fields are set
	if !c.init {
		c.initDefaults()
	}

	// parse the command arguments
	if len(args) > 1 {
		cmd, err := c.Flags().Parse(args[1:])
		if err != nil {
			return err
		}

		// if the command args contain a sub command, remove
		// the command from the args and execute the sub command
		if cmd != "" {
			if c, _ := c.GetCommand(cmd); c != nil {
				return c.execute(args[1:])
			}
			return ErrInvalidCommand
		}

		// if args contain a sub command flag, execute it
		// using the current parent command as the cmd, and
		// the current flags as the args
		for _, f := range c.Flags().Flags {
			if f.Type == CmdFlag && f.Bool() {
				return f.Cmd().Run(c, c.Flags())
			}
		}
	}

	// execute the command run function or
	// the command help function if the help flag is set
	if c.Run != nil {
		return c.Run(c, c.flags)
	}
	if h, _ := c.GetCommand("help"); h != nil {
		return h.Run(c, c.flags)
	}
	return ErrInvalidCommand
}

func (c *Command) GetCommand(name string) (*Command, int) {
	if c != nil {
		if cmd, i := c.Subs.GetIndex(name); cmd != nil {
			return cmd, i
		}

		// if the command is not found in the current command,
		// search the parent commands for the command if it
		// is a persistent command
		for _, p := range c.Parents {
			if c, _ := p.GetCommand(name); c != nil && c.Persist {
				return c, -1
			}
		}
	}
	return nil, -1
}

func (c *Command) AddCommand(cmds ...*Command) {
	if c == nil {
		return
	}
	for _, cmd := range cmds {
		c.Subs.Add(cmd)
		cmd.Parents.Add(c)
	}
}

func (c *Command) Flags() *FlagSet {
	if c == nil {
		return nil
	}
	if c.flags == nil {
		c.flags = NewFlagSet(c)
	}
	return c.flags
}

func (c *Command) String() string {
	if c == nil {
		return ""
	}
	c.initDefaults()
	b := buffer.Pool.Get()
	defer b.Free()
	b.MustWriteByte('\t')
	n := bytes.Repeat([]byte{' '}, 23)
	copy(n, c.Name)
	return b.MustWrite(n).
		MustWrite(c.Short.buf.Bytes()).
		String()
}

// --------------------------------------------------------------------------- /
// Commands definition
// --------------------------------------------------------------------------- /

// Commands is a slice of Command pointers.
type Commands []*Command

func NewCommands() Commands {
	return make([]*Command, 0)
}

func (c *Commands) Len() int {
	return len(*c)
}

func (c *Commands) GetIndex(name string) (*Command, int) {
	for i, cmd := range *c {
		if cmd.Name == name {
			return cmd, i
		}
	}
	return nil, -1
}

func (c *Commands) Get(name string) *Command {
	if c == nil {
		return nil
	}
	cmd, _ := c.GetIndex(name)
	return cmd
}

func (c *Commands) GetP(name string) *Command {
	cmd, _ := c.GetIndex(name)
	if cmd.Persist {
		return cmd
	}
	return nil
}

func (c *Commands) Add(cmd *Command) {
	if cmd == nil {
		return
	}
	if c == nil {
		*c = NewCommands()
	}
	if _, i := c.GetIndex(cmd.Name); i != -1 {
		(*c)[i] = cmd
		return
	}
	*c = append(*c, cmd)
}

func (c *Commands) String() string {
	if c.Len() == 0 {
		return ""
	}
	b := buffer.Pool.Get()
	defer b.Free()
	b.MustWriteString("Commands:").MustWrite(EOL)
	for _, cmd := range *c {
		b.MustWriteString(cmd.String()).MustWrite(EOL)
	}
	return b.String()
}
