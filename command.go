package structcli

import (
	"fmt"
	"reflect"
	"strings"
)

type commandSpec struct {
	name        string
	description string

	options       []*optionSpec
	optionsByName map[string]*optionSpec

	subcommands       []*commandSpec
	subcommandsByName map[string]*commandSpec

	positionals []*positionalSpec
}

func (c *commandSpec) extractStruct(v reflect.Value, visited map[reflect.Type]bool) error {
	t := v.Type()

	if visited[t] {
		return fmt.Errorf("command type `%s` is referenced recursively", t.String())
	}
	visited[t] = true

	return scanStruct(t, v, func(field *structField, value reflect.Value) error {
		if field.isCommand() {
			cmd, err := newCommandSpec(field)
			if err != nil {
				return err
			}

			err = c.validateSubcommand(cmd)
			if err != nil {
				return err
			}

			err = cmd.extractStruct(value.Elem(), visited)
			if err != nil {
				return err
			}

			c.addSubcommand(cmd)
			return nil
		}

		if field.isOption() {
			opt, err := newOptionSpec(field)
			if err != nil {
				return err
			}

			err = c.validateOption(opt)
			if err != nil {
				return err
			}

			c.addOption(opt)
			return nil
		}

		pos, err := newPositionalSpec(field)
		if err != nil {
			return err
		}

		if pos != nil {
			c.positionals = append(c.positionals, pos)
		}
		return nil
	})
}

func (c *commandSpec) validateSubcommand(cmd *commandSpec) error {
	if _, ok := c.subcommandsByName[cmd.name]; ok {
		return fmt.Errorf("duplicated subcommand name `%s`", cmd.name)
	}
	return nil
}

func (c *commandSpec) addSubcommand(cmd *commandSpec) {
	if c.subcommandsByName == nil {
		c.subcommandsByName = make(map[string]*commandSpec)
	}

	c.subcommandsByName[cmd.name] = cmd
	c.subcommands = append(c.subcommands, cmd)
}

func (c *commandSpec) validateOption(opt *optionSpec) error {
	names := opt.getNames()
	for _, name := range names {
		if _, ok := c.optionsByName[name]; ok {
			return fmt.Errorf("duplicated option name `%s`", name)
		}
	}
	return nil
}

func (c *commandSpec) addOption(opt *optionSpec) {
	if c.optionsByName == nil {
		c.optionsByName = make(map[string]*optionSpec)
	}

	names := opt.getNames()
	for _, name := range names {
		c.optionsByName[name] = opt
	}
}

func newCommandSpec(sf *structField) (*commandSpec, error) {
	name := sf.getCommandName()
	if strings.HasPrefix(name, "-") {
		return nil, fmt.Errorf("command name '%s' for field `%s.%s` cannot start with '-'", name, sf.Type.String(), sf.Name)
	}

	t, ptr := sf.indirectType()
	if !(t.Kind() == reflect.Struct && ptr) {
		return nil, fmt.Errorf("command `%s` must be a pointer to a struct", name)
	}

	return &commandSpec{
		name:        name,
		description: sf.getDescription(),
	}, nil
}
