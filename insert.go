package dbr

import (
	"bytes"
	"reflect"
)

// InsertStmt builds `INSERT INTO ...`
type InsertStmt struct {
	raw

	Table       string
	Column      []string
	Value       [][]interface{}
	UpsertValue map[string]interface{}
}

// Build builds `INSERT INTO ...` in dialect
func (b *InsertStmt) Build(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(d, buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("INSERT INTO ")
	buf.WriteString(d.QuoteIdent(b.Table))

	buf.WriteString(" (")

	placeholder := new(bytes.Buffer)
	placeholder.WriteRune('(')
	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(",")
			placeholder.WriteString(",")
		}
		buf.WriteString(d.QuoteIdent(col))
		placeholder.WriteString(d.Placeholder())
	}
	placeholder.WriteString(")")
	placeholderStr := placeholder.String()

	buf.WriteString(") VALUES ")

	for i, tuple := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(placeholderStr)

		buf.WriteValue(tuple...)
	}

	if len(b.UpsertValue) > 0 {
		buf.WriteString(" ON DUPLICATE KEY UPDATE ")

		i := 0
		for col, v := range b.UpsertValue {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(d.QuoteIdent(col))
			buf.WriteString(" = ")
			buf.WriteString(d.Placeholder())

			buf.WriteValue(v)
			i++
		}
	}

	return nil
}

// InsertInto creates an InsertStmt
func InsertInto(table string) *InsertStmt {
	return &InsertStmt{
		Table: table,
		UpsertValue: make(map[string]interface{}),

	}
}

// InsertBySql creates an InsertStmt from raw query
func InsertBySql(query string, value ...interface{}) *InsertStmt {
	return &InsertStmt{
		raw: raw{
			Query: query,
			Value: value,
			UpsertValue: make(map[string]interface{}),
		},
	}
}

// Columns adds columns
func (b *InsertStmt) Columns(column ...string) *InsertStmt {
	b.Column = column
	return b
}

// Values adds a tuple for columns
func (b *InsertStmt) Values(value ...interface{}) *InsertStmt {
	b.Value = append(b.Value, value)
	return b
}

// Record adds a tuple for columns from a struct
func (b *InsertStmt) Record(structValue interface{}) *InsertStmt {
	v := reflect.Indirect(reflect.ValueOf(structValue))

	if v.Kind() == reflect.Struct {
		var value []interface{}
		m := structMap(v)
		for _, key := range b.Column {
			if val, ok := m[key]; ok {
				value = append(value, val.Interface())
			} else {
				value = append(value, nil)
			}
		}
		b.Values(value...)
	}
	return b
}

func (b *InsertStmt) Set(column string, value interface{}) *InsertStmt {

	b.UpsertValue[column] = value
	return b
}

// SetMap specifies a list of key-value pair
func (b *InsertStmt) OnDuplicateKeyUpdate(m map[string]interface{}) *InsertStmt {
	for col, val := range m {
		b.Set(col, val)
	}
	return b
}
