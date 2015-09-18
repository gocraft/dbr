package dbr

import "errors"

// package errors
var (
	ErrNotFound           = errors.New("dbr: not found")
	ErrNotSupported       = errors.New("dbr: not supported")
	ErrTableNotSpecified  = errors.New("dbr: table not specified")
	ErrColumnNotSpecified = errors.New("dbr: column not specified")
	ErrLoadNonPointer     = errors.New("dbr: attempt to load into a non-pointer")
	ErrPlaceholderCount   = errors.New("dbr: wrong placeholder count")
)
