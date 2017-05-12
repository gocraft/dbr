package dbr

import (
	"bytes"
	"fmt"
	"reflect"
)

// ConflictStmt is ` ON CONFLICT ...` part of InsertStmt
type ConflictStmt struct {
	constraint string
	actions    map[string]interface{}
}

// InsertStmt builds `INSERT INTO ...`
type InsertStmt struct {
	raw

	Table    string
	Column   []string
	Value    [][]interface{}
	Conflict *ConflictStmt
}

// Proposed is reference to proposed value in on conflict clause
func Proposed(column string) Builder {
	return BuildFunc(func(d Dialect, b Buffer) error {
		_, err := b.WriteString(d.Proposed(column))
		return err
	})
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

	placeholderBuf := new(bytes.Buffer)
	placeholderBuf.WriteString("(")
	buf.WriteString(" (")
	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(",")
			placeholderBuf.WriteString(",")
		}
		buf.WriteString(d.QuoteIdent(col))
		placeholderBuf.WriteString(placeholder)
	}
	buf.WriteString(") VALUES ")
	placeholderBuf.WriteString(")")
	placeholderStr := placeholderBuf.String()

	for i, tuple := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(placeholderStr)

		buf.WriteValue(tuple...)
	}
	if b.Conflict != nil && len(b.Conflict.actions) > 0 {
		keyword := d.OnConflict(b.Conflict.constraint)
		if len(keyword) == 0 {
			return fmt.Errorf("Dialect %s does not support OnConflict", d)
		}
		buf.WriteString(" ")
		buf.WriteString(keyword)
		buf.WriteString(" ")
		needComma := false
		for _, column := range b.Column {
			if v, ok := b.Conflict.actions[column]; ok {
				if needComma {
					buf.WriteString(",")
				}
				buf.WriteString(d.QuoteIdent(column))
				buf.WriteString("=")
				buf.WriteString(placeholder)
				buf.WriteValue(v)
				needComma = true
			}
		}
	}

	return nil
}

// InsertInto creates an InsertStmt
func InsertInto(table string) *InsertStmt {
	return &InsertStmt{
		Table: table,
	}
}

// InsertBySql creates an InsertStmt from raw query
func InsertBySql(query string, value ...interface{}) *InsertStmt {
	return &InsertStmt{
		raw: raw{
			Query: query,
			Value: value,
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
		m := structMap(v.Type())
		for _, key := range b.Column {
			if index, ok := m[key]; ok {
				value = append(value, v.FieldByIndex(index).Interface())
			} else {
				value = append(value, nil)
			}
		}
		b.Values(value...)
	}
	return b
}

// OnConflictMap allows to add actions for constraint violation, e.g UPSERT
func (b *InsertStmt) OnConflictMap(constraint string, actions map[string]interface{}) *InsertStmt {
	b.Conflict = &ConflictStmt{constraint: constraint, actions: actions}
	return b
}

// OnConflict creates an empty OnConflict section fo insert statement , e.g UPSERT
func (b *InsertStmt) OnConflict(constraint string) *ConflictStmt {
	return b.OnConflictMap(constraint, make(map[string]interface{})).Conflict
}

// Action adds action for column which will do if conflict happens
func (b *ConflictStmt) Action(column string, action interface{}) *ConflictStmt {
	b.actions[column] = action
	return b
}
