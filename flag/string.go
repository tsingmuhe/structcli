package flag

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) String() string { return string(*s) }

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (f *FlagSet) StringVar(p *string, value string, shorthand, name, description string) {
	f.Var(newStringValue(value, p), shorthand, name, description)
}
