package dbr

import (
	"errors"
)

var (
	ErrNoTable  = errors.New("No Table specified")
	ErrNoValues = errors.New("No Values specified")
)
