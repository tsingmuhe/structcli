package structcli

import (
	"fmt"
	"reflect"
)

type CommandLine[T any] struct {
	name        string
	description string
	version     string

	command     *T
	commandSpec *commandSpec
}

func (c *CommandLine[T]) Parse(args []string) (*T, error) {
	err := c.parseCommandSpec()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *CommandLine[T]) parseCommandSpec() error {
	rv := reflect.ValueOf(c.command)
	if rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("command must be a struct, got %s", rv.Type().String())
	}

	if rv.IsNil() {
		return fmt.Errorf("command must not be nil")
	}

	spec := &commandSpec{
		name:        c.name,
		description: c.description,
	}

	err := spec.extractStruct(rv.Elem(), make(map[reflect.Type]bool))
	if err != nil {
		return err
	}

	c.commandSpec = spec
	return nil
}

func Create[T any](name, description, version string, command *T) *CommandLine[T] {
	return &CommandLine[T]{
		name:        name,
		description: description,
		version:     version,
		command:     command,
	}
}
