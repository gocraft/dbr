package dbr

import (
	"bytes"
	"database/sql"
	"reflect"
	"time"
)

type InsertBuilder struct {
	*Session
	runner

	Into string
	Cols []string
	Vals [][]interface{}
	Recs []interface{}
}

func (sess *Session) InsertInto(into string) *InsertBuilder {
	return &InsertBuilder{
		Session: sess,
		runner:  sess.cxn.Db,
		Into:    into,
	}
}

func (tx *Tx) InsertInto(into string) *InsertBuilder {
	return &InsertBuilder{
		Session: tx.Session,
		runner:  tx.Tx,
		Into:    into,
	}
}

func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.Cols = columns
	return b
}

func (b *InsertBuilder) Values(vals ...interface{}) *InsertBuilder {
	b.Vals = append(b.Vals, vals)
	return b
}

// Pulls in values to match Columns from the record
func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	b.Recs = append(b.Recs, record)
	return b
}

func (b *InsertBuilder) ToSql() (string, []interface{}) {
	if len(b.Into) == 0 {
		panic("no table specified")
	}
	if len(b.Cols) == 0 {
		panic("no columns specified")
	}
	if len(b.Vals) == 0 && len(b.Recs) == 0 {
		panic("no values or records specified")
	}

	var sql bytes.Buffer
	var placeholder bytes.Buffer // Build the placeholder like "(?,?,?)"
	var args []interface{}

	sql.WriteString("INSERT INTO ")
	sql.WriteString(b.Into)
	sql.WriteString(" (")

	// Simulataneously write the cols to the sql buffer, and build a placeholder
	placeholder.WriteRune('(')
	for i, c := range b.Cols {
		if i > 0 {
			sql.WriteRune(',')
			placeholder.WriteRune(',')
		}
		sql.WriteString(Quoter.QuoteColumn(c))
		placeholder.WriteRune('?')
	}
	sql.WriteString(") VALUES ")
	placeholder.WriteRune(')')
	placeholderStr := placeholder.String()

	// Go thru each value we want to insert. Write the placeholders, and collect args
	for i, row := range b.Vals {
		if i > 0 {
			sql.WriteRune(',')
		}
		sql.WriteString(placeholderStr)

		for _, v := range row {
			args = append(args, v)
		}
	}
	anyVals := len(b.Vals) > 0

	// Go thru the records. Write the placeholders, and do reflection on the records to extract args
	for i, rec := range b.Recs {
		if i > 0 || anyVals {
			sql.WriteRune(',')
		}
		sql.WriteString(placeholderStr)

		ind := reflect.Indirect(reflect.ValueOf(rec))
		vals, err := b.valuesFor(ind.Type(), ind, b.Cols)
		if err != nil {
			panic(err.Error())
		}
		for _, v := range vals {
			args = append(args, v)
		}
	}

	return sql.String(), args
}

// Executes the insert against the database
func (b *InsertBuilder) Exec() (sql.Result, error) {
	sql, args := b.ToSql()

	fullSql, err := Interpolate(sql, args)
	if err != nil {
		return nil, b.EventErrKv("dbr.insert.exec.interpolate", err, kvs{"sql": fullSql})
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { b.TimingKv("dbr.insert", time.Since(startTime).Nanoseconds(), kvs{"sql": fullSql}) }()

	result, err := b.runner.Exec(fullSql)
	if err != nil {
		return result, b.EventErrKv("dbr.insert.exec.exec", err, kvs{"sql": fullSql})
	}

	// If the structure has an "Id" field which is an int64, set it from the LastInsertId(). Otherwise, don't bother.
	if len(b.Recs) == 1 {
		rec := b.Recs[0]
		val := reflect.Indirect(reflect.ValueOf(rec))
		if val.Kind() == reflect.Struct && val.CanSet() {
			if idField := val.FieldByName("Id"); idField.IsValid() && idField.Kind() == reflect.Int64 {
				if lastId, err := result.LastInsertId(); err == nil {
					idField.Set(reflect.ValueOf(lastId))
				} else {
					b.EventErrKv("dbr.insert.exec.last_inserted_id", err, kvs{"sql": fullSql})
				}
			}
		}
	}

	return result, nil
}
