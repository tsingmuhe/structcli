package flag

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Flag struct {
	ShortName   string
	Name        string
	Description string
	DefValue    string // default value (as text); for usage message

	Value Value
}

type Value interface {
	String() string
	Set(string) error
	IsBool() bool
}

type FlagSet struct {
	name string

	formal     map[string]*Flag
	shorthands map[byte]*Flag

	parsed bool
	actual map[string]*Flag
	args   []string
}

func (f *FlagSet) BoolVar(p *bool, value bool, short, name, description string) {
	f.Var(newBoolValue(value, p), short, name, description)
}

func (f *FlagSet) IntVar(p *int, value int, short, name, description string) {
	f.Var(newIntValue(value, p), short, name, description)
}

func (f *FlagSet) Int64Var(p *int64, value int64, short, name, description string) {
	f.Var(newInt64Value(value, p), short, name, description)
}

func (f *FlagSet) UintVar(p *uint, value uint, short, name, description string) {
	f.Var(newUintValue(value, p), short, name, description)
}

func (f *FlagSet) Uint64Var(p *uint64, value uint64, short, name, description string) {
	f.Var(newUint64Value(value, p), short, name, description)
}

func (f *FlagSet) Float64Var(p *float64, value float64, short, name, description string) {
	f.Var(newFloat64Value(value, p), short, name, description)
}

func (f *FlagSet) StringVar(p *string, value string, short, name, description string) {
	f.Var(newStringValue(value, p), short, name, description)
}

func (f *FlagSet) TextVar(p encoding.TextUnmarshaler, value encoding.TextMarshaler, short, name, description string) {
	f.Var(newTextValue(value, p), short, name, description)
}

func (f *FlagSet) Var(value Value, short, name, description string) {
	// Flag must not begin "-" or contain "=".
	if len(short) > 1 {
		panic(f.sprintf("flag %q is more than one ASCII character", short))
	}

	if len(short) > 0 {
		if strings.HasPrefix(short, "-") {
			panic(f.sprintf("flag %q begins with -", short))
		} else if strings.Contains(short, "=") {
			panic(f.sprintf("flag %q contains =", short))
		}
	}

	if strings.HasPrefix(name, "-") {
		panic(f.sprintf("flag %q begins with -", name))
	} else if strings.Contains(name, "=") {
		panic(f.sprintf("flag %q contains =", name))
	}

	f.addFlag(&Flag{
		ShortName:   short,
		Name:        name,
		Description: description,
		DefValue:    value.String(),
	})
}

func (f *FlagSet) addFlag(flag *Flag) {
	_, alreadythere := f.formal[flag.Name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = f.sprintf("flag redefined: %s", flag.Name)
		} else {
			msg = f.sprintf("%s flag redefined: %s", f.name, flag.Name)
		}
		panic(msg)
	}

	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}
	f.formal[flag.Name] = flag

	if flag.ShortName == "" {
		return
	}

	short := flag.ShortName[0]
	_, alreadythere = f.shorthands[short]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = f.sprintf("flag redefined: %s", short)
		} else {
			msg = f.sprintf("%s flag redefined: %s", f.name, short)
		}
		panic(msg)
	}

	if f.shorthands == nil {
		f.shorthands = make(map[byte]*Flag)
	}
	f.shorthands[short] = flag
}

func (f *FlagSet) Parse(args []string) error {
	f.parsed = true
	f.args = make([]string, 0, len(args))

	if len(args) == 0 {
		return nil
	}

	err := f.parse(args)
	if err != nil {
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			if err == ErrHelp {
				os.Exit(0)
			}
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}

	return nil
}

func (f *FlagSet) parse(args []string) (err error) {
	for len(args) > 0 {
		s := args[0]
		args = args[1:]

		if len(s) < 2 || s[0] != '-' {
			f.args = append(f.args, s)
			continue
		}

		if s[1] == '-' {
			if len(s) == 2 {
				f.args = append(f.args, args...)
				break
			}
			args, err = f.parseOneLong(s, args)
		} else {
			args, err = f.parseShort(s, args)
		}

		if err != nil {
			return
		}
	}

	return
}

func (f *FlagSet) parseOneLong(s string, args []string) ([]string, error) {
	name := s[2:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return args, f.failf("bad flag syntax: %s", s)
	}

	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ {
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	flag, ok := f.formal[name]
	if !ok {
		return args, f.failf("flag provided but not defined: -%s", name)
	}

	if flag.Value.IsBool() { // special case: doesn't need an arg
		if hasValue { // '--flag=arg'
			if err := flag.Value.Set(value); err != nil {
				return args, f.failf("invalid boolean value %q for -%s: %v", value, name, err)
			}
		} else { // '--flag'
			if err := flag.Value.Set("true"); err != nil {
				return args, f.failf("invalid boolean flag %s: %v", name, err)
			}
		}
	} else {
		// It must have a value, which might be the next argument.
		if !hasValue && len(args) > 0 { // '--flag arg'
			// value is the next arg
			hasValue = true
			value, args = args[0], args[1:]
		}

		if !hasValue { //'--flag'
			return args, f.failf("flag needs an argument: -%s", name)
		}

		if err := flag.Value.Set(value); err != nil {
			return args, f.failf("invalid value %q for flag -%s: %v", value, name, err)
		}
	}

	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}

	f.actual[name] = flag
	return args, nil
}

func (f *FlagSet) parseShort(s string, args []string) ([]string, error) {
	shorthands := s[1:]
	if len(shorthands) == 0 || shorthands[0] == '-' || shorthands[0] == '=' {
		return args, f.failf("bad flag syntax: %s", s)
	}

	for len(shorthands) > 0 {
		shorthands, a, err := f.parseOneShort(shorthands, args)
		if err != nil {
			return a, err
		}
	}

	return
}

func (f *FlagSet) parseOneShort(s string, args []string) ([]string, []string, error) {

}

func (f *FlagSet) Lookup(name string) *Flag {
	return nil
}

func (f *FlagSet) NArg() int { return len(f.args) }

func (f *FlagSet) Args() []string { return f.args }

func (f *FlagSet) failf(format string, a ...any) error {
	msg := f.sprintf(format, a...)
	return errors.New(msg)
}

func (f *FlagSet) sprintf(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(os.Stderr, msg)
	return msg
}
