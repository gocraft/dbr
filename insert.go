package dbr

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
)

// InsertStmt builds `INSERT INTO ...`
type InsertStmt struct {
	runner
	EventReceiver
	Dialect

	raw

	Table        string
	Column       []string
	Value        [][]interface{}
	ReturnColumn []string
	RecordID     *int64
}

type InsertBuilder = InsertStmt

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

	var placeholderBuf strings.Builder
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

	if len(b.ReturnColumn) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.ReturnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(d.QuoteIdent(col))
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

func (sess *Session) InsertInto(table string) *InsertStmt {
	b := InsertInto(table)
	b.runner = sess
	b.EventReceiver = sess
	b.Dialect = sess.Dialect
	return b
}

func (tx *Tx) InsertInto(table string) *InsertStmt {
	b := InsertInto(table)
	b.runner = tx
	b.EventReceiver = tx
	b.Dialect = tx.Dialect
	return b
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

func (sess *Session) InsertBySql(query string, value ...interface{}) *InsertStmt {
	b := InsertBySql(query, value...)
	b.runner = sess
	b.EventReceiver = sess
	b.Dialect = sess.Dialect
	return b
}

func (tx *Tx) InsertBySql(query string, value ...interface{}) *InsertStmt {
	b := InsertBySql(query, value...)
	b.runner = tx
	b.EventReceiver = tx
	b.Dialect = tx.Dialect
	return b
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
		m := structMap(v)
		if v.CanSet() {
			// ID is recommended by golint here
			if field, ok := m["id"]; ok && field.Kind() == reflect.Int64 {
				b.RecordID = field.Addr().Interface().(*int64)
			}
		}

		var value []interface{}
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

func (b *InsertStmt) Returning(column ...string) *InsertStmt {
	b.ReturnColumn = column
	return b
}

func (b *InsertStmt) Pair(column string, value interface{}) *InsertStmt {
	b.Column = append(b.Column, column)
	switch len(b.Value) {
	case 0:
		b.Values(value)
	case 1:
		b.Value[0] = append(b.Value[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}

func (b *InsertStmt) Exec() (sql.Result, error) {
	return b.ExecContext(context.Background())
}

func (b *InsertStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	result, err := exec(ctx, b.runner, b.EventReceiver, b, b.Dialect)
	if err != nil {
		return nil, err
	}

	if b.RecordID != nil {
		if id, err := result.LastInsertId(); err == nil {
			*b.RecordID = id
		}
		b.RecordID = nil
	}

	return result, nil
}

func (b *InsertStmt) LoadContext(ctx context.Context, value interface{}) error {
	_, err := query(ctx, b.runner, b.EventReceiver, b, b.Dialect, value)
	return err
}

func (b *InsertStmt) Load(value interface{}) error {
	return b.LoadContext(context.Background(), value)
}
