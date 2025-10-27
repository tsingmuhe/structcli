package flag

import "strconv"

type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (f *FlagSet) Int64Var(p *int64, value int64, shorthand, name, description string) {
	f.Var(newInt64Value(value, p), shorthand, name, description)
}
