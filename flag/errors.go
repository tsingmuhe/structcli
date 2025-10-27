package flag

import (
	"errors"
	"strconv"
)

var errParse = errors.New("parse error")
var errRange = errors.New("value out of range")

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}

	if ne.Err == strconv.ErrSyntax {
		return errParse
	}

	if ne.Err == strconv.ErrRange {
		return errRange
	}

	return err
}
