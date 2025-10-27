package flag

import (
	"encoding"
	"fmt"
	"reflect"
)

type textValue struct{ p encoding.TextUnmarshaler }

func newTextValue(val encoding.TextMarshaler, p encoding.TextUnmarshaler) textValue {
	ptrVal := reflect.ValueOf(p)
	if ptrVal.Kind() != reflect.Ptr {
		panic("variable value type must be a pointer")
	}

	defVal := reflect.ValueOf(val)
	if defVal.Kind() == reflect.Ptr {
		defVal = defVal.Elem()
	}

	if defVal.Type() != ptrVal.Type().Elem() {
		panic(fmt.Sprintf("default type does not match variable type: %v != %v", defVal.Type(), ptrVal.Type().Elem()))
	}

	ptrVal.Elem().Set(defVal)
	return textValue{p}
}

func (v textValue) String() string {
	if m, ok := v.p.(encoding.TextMarshaler); ok {
		if b, err := m.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}

func (v textValue) Set(s string) error {
	return v.p.UnmarshalText([]byte(s))
}

func (f *FlagSet) TextVar(p encoding.TextUnmarshaler, value encoding.TextMarshaler, shorthand, name, description string) {
	f.Var(newTextValue(value, p), shorthand, name, description)
}
