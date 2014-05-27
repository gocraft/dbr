package dbr

import (
	"bytes"
	"fmt"
)

type SelectBuilder struct {
	*Session
	runner

	RawFullSql   string
	RawArguments []interface{}

	IsDistinct      bool
	Columns         []string
	FromTable       string
	WhereFragments  []*whereFragment
	GroupBys        []string
	HavingFragments []*whereFragment
	OrderBys        []string
	LimitCount      uint64
	LimitValid      bool
	OffsetCount     uint64
	OffsetValid     bool
}

func (sess *Session) Select(cols ...string) *SelectBuilder {
	return &SelectBuilder{
		Session: sess,
		runner:  sess.cxn.Db,
		Columns: cols,
	}
}

func (sess *Session) SelectBySql(sql string, args ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		Session:      sess,
		runner:       sess.cxn.Db,
		RawFullSql:   sql,
		RawArguments: args,
	}
}

func (tx *Tx) Select(cols ...string) *SelectBuilder {
	return &SelectBuilder{
		Session: tx.Session,
		runner:  tx.Tx,
		Columns: cols,
	}
}

func (tx *Tx) SelectBySql(sql string, args ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		Session:      tx.Session,
		runner:       tx.Tx,
		RawFullSql:   sql,
		RawArguments: args,
	}
}

func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.IsDistinct = true
	return b
}

func (b *SelectBuilder) From(from string) *SelectBuilder {
	b.FromTable = from
	return b
}

func (b *SelectBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

func (b *SelectBuilder) GroupBy(group string) *SelectBuilder {
	b.GroupBys = append(b.GroupBys, group)
	return b
}

func (b *SelectBuilder) Having(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.HavingFragments = append(b.HavingFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

func (b *SelectBuilder) OrderBy(ord string) *SelectBuilder {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

func (b *SelectBuilder) OrderDir(ord string, isAsc bool) *SelectBuilder {
	if isAsc {
		b.OrderBys = append(b.OrderBys, ord+" ASC")
	} else {
		b.OrderBys = append(b.OrderBys, ord+" DESC")
	}
	return b
}

func (b *SelectBuilder) Limit(limit uint64) *SelectBuilder {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

func (b *SelectBuilder) Offset(offset uint64) *SelectBuilder {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

// Assumes page/perPage are valid. Page and perPage must be >= 1
func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

func (b *SelectBuilder) ToSql() (string, []interface{}) {
	if b.RawFullSql != "" {
		return b.RawFullSql, b.RawArguments
	}

	if len(b.Columns) == 0 {
		panic("no columns specified")
	}
	if len(b.FromTable) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("SELECT ")

	if b.IsDistinct {
		sql.WriteString("DISTINCT ")
	}

	for i, s := range b.Columns {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(s)
	}

	sql.WriteString(" FROM ")
	sql.WriteString(b.FromTable)

	if len(b.WhereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.WhereFragments, &sql, &args)
	}

	if len(b.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		for i, s := range b.GroupBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if len(b.HavingFragments) > 0 {
		sql.WriteString(" HAVING ")
		writeWhereFragmentsToSql(b.HavingFragments, &sql, &args)
	}

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
