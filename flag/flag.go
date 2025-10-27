package flag

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Flag struct {
	Shorthand   string
	Name        string
	Description string
	DefValue    string
	Value       Value
}

type Value interface {
	String() string
	Set(string) error
}

type BoolValue interface {
	Value
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

func (f *FlagSet) Var(value Value, shorthand, name, description string) {
	if len(shorthand) > 1 {
		panic(f.sprintf("flag shorthand `%s` is more than one ASCII character", shorthand))
	}

	if len(shorthand) == 1 {
		if shorthand == "-" {
			panic(f.sprintf("flag shorthand `%s` begins with -", shorthand))
		} else if shorthand == "=" {
			panic(f.sprintf("flag shorthand `%s` contains =", shorthand))
		}
	}

	// Flag must not begin "-" or contain "=".
	if strings.HasPrefix(name, "-") {
		panic(f.sprintf("flag `%s` begins with -", name))
	} else if strings.Contains(name, "=") {
		panic(f.sprintf("flag `%s` contains =", name))
	}

	f.addFlag(&Flag{
		Shorthand:   shorthand,
		Name:        name,
		Description: description,
		DefValue:    value.String(),
	})
}

func (f *FlagSet) addFlag(flag *Flag) {
	_, alreadyThere := f.formal[flag.Name]
	if alreadyThere {
		var msg string
		if f.name == "" {
			msg = f.sprintf("flag redefined: --%s", flag.Name)
		} else {
			msg = f.sprintf("%s flag redefined: --%s", f.name, flag.Name)
		}
		panic(msg)
	}

	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}

	f.formal[flag.Name] = flag

	if flag.Shorthand == "" {
		return
	}

	short := flag.Shorthand[0]
	_, alreadyThere = f.shorthands[short]
	if alreadyThere {
		var msg string
		if f.name == "" {
			msg = f.sprintf("flag redefined: -%s", flag.Shorthand)
		} else {
			msg = f.sprintf("%s flag redefined: -%s", f.name, flag.Shorthand)
		}
		panic(msg)
	}

	if f.shorthands == nil {
		f.shorthands = make(map[byte]*Flag)
	}

	f.shorthands[short] = flag
}

func (f *FlagSet) Parse(args []string) (err error) {
	f.parsed = true
	f.args = make([]string, 0, len(args))

	for len(args) > 0 {
		arg0 := args[0]
		args = args[1:]

		if len(arg0) < 2 || arg0[0] != '-' {
			f.args = append(f.args, arg0)
			continue
		}

		if arg0[1] == '-' {
			if len(arg0) == 2 {
				f.args = append(f.args, args...)
				break
			}

			args, err = f.parseLong(arg0, args)
		} else {
			args, err = f.parseShort(arg0, args)
		}

		if err != nil {
			return
		}
	}

	return
}

func (f *FlagSet) parseLong(arg0 string, args []string) ([]string, error) {
	name := arg0[2:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return args, f.failf("bad flag syntax: %s", arg0)
	}

	value, hasValue := "", false
	if idx := strings.IndexByte(name, '='); idx != -1 {
		name, value, hasValue = name[:idx], name[idx+1:], true
	}

	flag, ok := f.formal[name]
	if !ok {
		return args, f.failf("flag provided but not defined: --%s", name)
	}

	if fv, ok := flag.Value.(BoolValue); ok && fv.IsBool() {
		if hasValue {
			if err := flag.Value.Set(value); err != nil {
				return args, f.failf("invalid boolean value %q for --%s", value, name)
			}
		} else {
			if err := flag.Value.Set("true"); err != nil {
				return args, f.failf("invalid boolean flag: --%s", name)
			}
		}
	} else {
		if !hasValue && len(args) > 0 {
			hasValue = true
			value, args = args[0], args[1:]
		}

		if !hasValue {
			return args, f.failf("flag needs an argument: --%s", name)
		}

		if err := flag.Value.Set(value); err != nil {
			return args, f.failf("invalid value %q for flag --%s", value, name)
		}
	}

	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}

	f.actual[name] = flag
	return args, nil
}

func (f *FlagSet) parseShort(s string, args []string) (a []string, err error) {
	name := s[1:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return args, f.failf("bad flag syntax: %s", s)
	}

	value, hasValue := "", false
	if idx := strings.IndexByte(name, '='); idx != -1 {
		name, value, hasValue = name[:idx], name[idx+1:], true
	}

	for i := 0; i < len(name); i++ {
		short := name[i]
		flag, ok := f.shorthands[short]
		if !ok {
			return args, f.failf("flag provided but not defined: -%v", short)
		}

		isLast := i == len(name)-1
		if isLast { //last one
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
		} else {
			if flag.Value.IsBool() {
				if err := flag.Value.Set("true"); err != nil {
					return args, f.failf("invalid boolean flag %s: %v", name, err)
				}
			} else {
				return args, f.failf("flag needs an argument: -%s", name)
			}
		}
	}

	return
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
