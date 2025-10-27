package structcli

import (
	"fmt"
	"reflect"
)

type positionalSpec struct {
	description string

	placeholder string
	required    bool

	values  []string
	changed bool
}

func newPositionalSpec(sf *structField) (*positionalSpec, error) {
	t, ptr := sf.indirectType()

	switch {
	case t.Kind() == reflect.String:
		return &positionalSpec{}, nil
	case t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.String:
		return &positionalSpec{
			description: "",
			placeholder: "",
			required:    false,
			values:      nil,
			changed:     false,
		}, nil
	default:
		return nil, fmt.Errorf("option field `%s` type is invalid: %s", sf.Name, t.String())
	}
}
