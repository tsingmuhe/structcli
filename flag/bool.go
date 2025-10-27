package flag

import "strconv"

type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) IsBool() bool { return true }

func (f *FlagSet) BoolVar(p *bool, value bool, shorthand, name, description string) {
	f.Var(newBoolValue(value, p), shorthand, name, description)
}
