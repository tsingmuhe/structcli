package flag

import "strconv"

type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (f *FlagSet) Uint64Var(p *uint64, value uint64, shorthand, name, description string) {
	f.Var(newUint64Value(value, p), shorthand, name, description)
}
