package dbr

import "bytes"

// Buffer is an interface used by Builder to store intermediate results
type Buffer interface {
	WriteString(s string) (n int, err error)
	String() string

	WriteValue(v ...interface{}) (err error)
	Value() []interface{}
}

func newBuffer() Buffer {
	return &buffer{}
}

type buffer struct {
	bytes.Buffer
	v []interface{}
}

func (b *buffer) WriteValue(v ...interface{}) error {
	b.v = append(b.v, v...)
	return nil
}

func (b *buffer) Value() []interface{} {
	return b.v
}
