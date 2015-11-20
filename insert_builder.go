package dbr

import (
	"database/sql"
	"reflect"

	"github.com/gocraft/dbr/dialect"
)

type InsertBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	recordID       reflect.Value
	recordIDColumn string // postgres RETURNING

	*InsertStmt
}

func (sess *Session) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		InsertStmt:    InsertInto(table),
	}
}

func (tx *Tx) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		InsertStmt:    InsertInto(table),
	}
}

func (sess *Session) InsertBySql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		InsertStmt:    InsertBySql(query, value...),
	}
}

func (tx *Tx) InsertBySql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		InsertStmt:    InsertBySql(query, value...),
	}
}

func (b *InsertBuilder) ToSql() (string, []interface{}) {
	buf := NewBuffer()
	err := b.Build(b.Dialect, buf)
	if err != nil {
		panic(err)
	}
	if b.recordIDColumn != "" {
		buf.WriteString(" RETURNING ")
		buf.WriteString(b.Dialect.QuoteIdent(b.recordIDColumn))
	}
	return buf.String(), buf.Value()
}

func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.Column = append(b.Column, column)
	switch len(b.Value) {
	case 0:
		b.InsertStmt.Values(value)
	case 1:
		b.Value[0] = append(b.Value[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}

type sqlResult struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *sqlResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *sqlResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (b *InsertBuilder) Exec() (sql.Result, error) {
	if len(b.Value) > 1 {
		// remove reference to object for multiple insert
		b.recordID = reflect.Value{}
		b.recordIDColumn = ""
	}

	var (
		result sql.Result
		err    error
	)
	if b.recordIDColumn != "" {
		// for postgres, try to load the returning id
		var lastInsertId int64
		_, err = query(b.runner, b.EventReceiver, b, b.Dialect, &lastInsertId)
		if err == nil {
			result = &sqlResult{
				// implied because we only allow id injection if record count == 1
				rowsAffected: 1,
				lastInsertId: lastInsertId,
			}
		}
	} else {
		// normal
		result, err = exec(b.runner, b.EventReceiver, b, b.Dialect)
	}
	if err != nil {
		return nil, err
	}

	if b.recordID.IsValid() {
		if id, err := result.LastInsertId(); err == nil {
			b.recordID.SetInt(id)
		}
	}

	return result, nil
}

func (b *InsertBuilder) Columns(column ...string) *InsertBuilder {
	b.InsertStmt.Columns(column...)
	return b
}

func (b *InsertBuilder) Record(structValue interface{}) *InsertBuilder {
	v := reflect.Indirect(reflect.ValueOf(structValue))
	if v.Kind() == reflect.Struct && v.CanSet() {
		// ID is recommended by golint here
		for _, name := range []string{"Id", "ID"} {
			fieldValue := v.FieldByName(name)
			if fieldValue.IsValid() && fieldValue.Kind() == reflect.Int64 {
				if b.Dialect == dialect.PostgreSQL {
					field, _ := v.Type().FieldByName(name)
					col := columnName(field)
					if col == "" {
						continue
					}
					b.recordIDColumn = col
				}
				b.recordID = fieldValue
				break
			}
		}
	}

	b.InsertStmt.Record(structValue)
	return b
}

func (b *InsertBuilder) Values(value ...interface{}) *InsertBuilder {
	b.InsertStmt.Values(value...)
	return b
}
