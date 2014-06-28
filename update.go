package dbr

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

type UpdateBuilder struct {
	*Session
	runner

	RawFullSql   string
	RawArguments []interface{}

	Table          string
	SetClauses     []*setClause
	WhereFragments []*whereFragment
	OrderBys       []string
	LimitCount     uint64
	LimitValid     bool
	OffsetCount    uint64
	OffsetValid    bool
}

type setClause struct {
	column string
	value  interface{}
}

func (sess *Session) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		Session: sess,
		runner:  sess.cxn.Db,
		Table:   table,
	}
}

func (sess *Session) UpdateBySql(sql string, args ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		Session:      sess,
		runner:       sess.cxn.Db,
		RawFullSql:   sql,
		RawArguments: args,
	}
}

func (tx *Tx) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		Session: tx.Session,
		runner:  tx.Tx,
		Table:   table,
	}
}

func (tx *Tx) UpdateBySql(sql string, args ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		Session:      tx.Session,
		runner:       tx.Tx,
		RawFullSql:   sql,
		RawArguments: args,
	}
}

func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.SetClauses = append(b.SetClauses, &setClause{column: column, value: value})
	return b
}

func (b *UpdateBuilder) SetMap(clauses map[string]interface{}) *UpdateBuilder {
	for col, val := range clauses {
		b = b.Set(col, val)
	}
	return b
}

func (b *UpdateBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *UpdateBuilder {
	b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

func (b *UpdateBuilder) OrderBy(ord string) *UpdateBuilder {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

func (b *UpdateBuilder) OrderDir(ord string, isAsc bool) *UpdateBuilder {
	if isAsc {
		b.OrderBys = append(b.OrderBys, ord+" ASC")
	} else {
		b.OrderBys = append(b.OrderBys, ord+" DESC")
	}
	return b
}

func (b *UpdateBuilder) Limit(limit uint64) *UpdateBuilder {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

func (b *UpdateBuilder) Offset(offset uint64) *UpdateBuilder {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

func (b *UpdateBuilder) ToSql() (string, []interface{}) {
	if b.RawFullSql != "" {
		return b.RawFullSql, b.RawArguments
	}

	if len(b.Table) == 0 {
		panic("no table specified")
	}
	if len(b.SetClauses) == 0 {
		panic("no set clauses specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("UPDATE ")
	sql.WriteString(b.Table)
	sql.WriteString(" SET ")

	// Build SET clause SQL with placeholders and add values to args
	for i, c := range b.SetClauses {
		if i > 0 {
			sql.WriteString(", ")
		}
		Quoter.writeQuotedColumn(c.column, &sql)
		if e, ok := c.value.(*expr); ok {
			sql.WriteString(" = ")
			sql.WriteString(e.Sql)
			args = append(args, e.Values...)
		} else {
			sql.WriteString(" = ?")
			args = append(args, c.value)
		}
	}

	// Write WHERE clause if we have any fragments
	if len(b.WhereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.WhereFragments, &sql, &args)
	}

	// Ordering and limiting
	if len(b.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, s := range b.OrderBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if b.LimitValid {
		sql.WriteString(" LIMIT ")
		fmt.Fprint(&sql, b.LimitCount)
	}

	if b.OffsetValid {
		sql.WriteString(" OFFSET ")
		fmt.Fprint(&sql, b.OffsetCount)
	}

	return sql.String(), args
}

func (b *UpdateBuilder) Exec() (sql.Result, error) {
	sql, args := b.ToSql()

	fullSql, err := Interpolate(sql, args)
	if err != nil {
		return nil, b.EventErrKv("dbr.update.exec.interpolate", err, kvs{"sql": fullSql})
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { b.TimingKv("dbr.update", time.Since(startTime).Nanoseconds(), kvs{"sql": fullSql}) }()

	result, err := b.runner.Exec(fullSql)
	if err != nil {
		return result, b.EventErrKv("dbr.update.exec.exec", err, kvs{"sql": fullSql})
	}

	return result, nil
}
