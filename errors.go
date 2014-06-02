package dbr

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNotUtf8 = errors.New("invalid UTF-8")
	ErrInvalidSliceLength = errors.New("length of slice is 0. length must be >= 1")
	ErrInvalidSliceValue = errors.New("trying to interpolate invalid slice value into query")
	ErrInvalidValue = errors.New("trying to interpolate invalid value into query")
	ErrArgumentMismatch = errors.New("mismatch between ? (placeholders) and arguments")
)
