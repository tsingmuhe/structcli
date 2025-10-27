package flag

import "strconv"

type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (f *FlagSet) UintVar(p *uint, value uint, shorthand, name, description string) {
	f.Var(newUintValue(value, p), shorthand, name, description)
}
