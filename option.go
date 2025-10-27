package structcli

import (
	"fmt"
	"reflect"
	"strings"
)

type optionSpec struct {
	shortName string
	longName  string

	placeholder string
	description string
	required    bool

	isBool    bool
	negatable bool

	values  []string
	changed bool
}

func (o *optionSpec) getNames() []string {
	var names []string

	if len(o.shortName) > 0 {
		names = append(names, o.shortName)
	}

	if len(o.longName) > 0 {
		names = append(names, o.longName)
	}

	if o.negatable {
		cutName, _ := strings.CutPrefix(o.longName, "--")
		names = append(names, "--no-"+cutName)
	}

	return names
}

func newOptionSpec(sf *structField) (*optionSpec, error) {
	shortName := sf.getShortName()
	longName, negatable := sf.getLongName()

	t, ptr := sf.indirectType()

	switch t.Kind() {
	case reflect.Bool:
		return &optionSpec{
			shortName:   shortName,
			longName:    longName,
			placeholder: "",
			description: sf.getDescription(),
			required:    !ptr,
			isBool:      true,
			negatable:   negatable,
		}, nil
	case reflect.String:
		return &optionSpec{
			shortName:   shortName,
			longName:    longName,
			placeholder: sf.getPlaceholder(),
			description: sf.getDescription(),
			required:    !ptr,
			isBool:      false,
			negatable:   false,
		}, nil
	default:
		return nil, fmt.Errorf("option field `%s` type is invalid: %s", sf.Name, t.String())
	}
}
