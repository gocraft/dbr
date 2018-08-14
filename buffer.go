package dbr

import "strings"

// Buffer collects strings, and values that are ready to be interpolated.
// This is used internally to efficiently build SQL statement.
type Buffer interface {
	WriteString(string) (int, error)
	String() string

	WriteValue(v ...interface{}) (err error)
	Value() []interface{}
}

type buffer struct {
	strings.Builder
	v []interface{}
}

// NewBuffer creates a new Buffer.
func NewBuffer() Buffer {
	return &buffer{}
}

func (b *buffer) WriteValue(v ...interface{}) error {
	b.v = append(b.v, v...)
	return nil
}

func (b *buffer) Value() []interface{} {
	return b.v
}
