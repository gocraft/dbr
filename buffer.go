package dbr

import "bytes"

type StringWriter interface {
	WriteString(s string) (n int, err error)
	String() string
}

type Buffer interface {
	StringWriter

	WriteValue(v ...interface{}) (err error)
	Value() []interface{}
}

type buffer struct {
	StringWriter
	v []interface{}
}

func NewBuffer() Buffer {
	return &buffer{
		StringWriter: new(bytes.Buffer),
	}
}

func (b *buffer) WriteValue(v ...interface{}) error {
	b.v = append(b.v, v...)
	return nil
}

func (b *buffer) Value() []interface{} {
	return b.v
}
