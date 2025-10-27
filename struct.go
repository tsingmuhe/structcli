package structcli

import (
	"reflect"
	"strings"
)

type structField reflect.StructField

func (s *structField) indirectType() (reflect.Type, bool) {
	if s.Type.Kind() != reflect.Pointer {
		return s.Type, false
	}
	return s.Type.Elem(), true
}

func (s *structField) isCommand() bool {
	if s.Tag.Get("command") == "" {
		return false
	}
	return true
}

func (s *structField) getCommandName() string {
	return s.Tag.Get("command")
}

func (s *structField) isOption() bool {
	if len(s.Tag.Get("short")) == 0 && len(s.Tag.Get("long")) == 0 {
		return false
	}
	return true
}

func (s *structField) getShortName() string {
	return s.Tag.Get("short")
}

func (s *structField) getLongName() (string, bool) {
	tagParts := strings.Split(s.Tag.Get("long"), ",")
	if len(tagParts) == 0 {
		return "", false
	}

	if len(tagParts) == 1 {
		return tagParts[0], false
	}

	negatable := false
	for _, val := range tagParts[1:] {
		if val == "negatable" {
			negatable = true
		}
	}

	return tagParts[0], negatable
}

func (s *structField) getPlaceholder() string {
	placeholder := s.Tag.Get("placeholder")
	if placeholder == "" {
		placeholder = s.Name
	}
	return placeholder
}

func (s *structField) getDescription() string {
	return s.Tag.Get("description")
}

type scanHandler func(*structField, reflect.Value) error

func scanStruct(t reflect.Type, v reflect.Value, handler scanHandler) error {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		if !sf.IsExported() {
			// Ignore unexported fields.
			continue
		}

		if sf.Anonymous {
			t := sf.Type
			if t.Kind() != reflect.Pointer {
				continue
			}

			if t.Elem().Kind() != reflect.Struct {
				continue
			}
		}

		f := structField(sf)
		err := handler(&f, v.Field(i))
		if err != nil {
			return err
		}
	}

	return nil
}
