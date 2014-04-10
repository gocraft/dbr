package dbr

import (
	"fmt"
	"strings"
)

type Query struct {
	*Session

	SelectSql      string
	FromSql        string
	WhereFragments []string
	Params         []interface{}
	OrderSql       string
	LimitCount     uint64
	LimitValid     bool
	OffsetCount    uint64
	OffsetValid    bool
}

func (q *Query) From(fromSql string) *Query {
	q.FromSql = fromSql
	return q
}

func (q *Query) Where(fromSql string, args ...interface{}) *Query {
	q.WhereFragments = append(q.WhereFragments, fromSql)
	for _, a := range args {
		q.Params = append(q.Params, a)
	}
	return q
}

func (q *Query) Limit(limit uint64) *Query {
	q.LimitCount = limit
	q.LimitValid = true
	return q
}

func (q *Query) Offset(offset uint64) *Query {
	q.OffsetCount = offset
	q.OffsetValid = true
	return q
}

// Assumes page/perPage are valid. Page and perPage must be >= 1
func (q *Query) Paginate(page, perPage uint64) *Query {
	q.Limit(perPage)
	q.Offset((page - 1) * perPage)
	return q
}

func (q *Query) Order(ord string) *Query {
	q.OrderSql = ord
	return q
}

func (q *Query) OrderDir(ord string, isAsc bool) *Query {
	if isAsc {
		q.OrderSql = ord + " ASC"
	} else {
		q.OrderSql = ord + " DESC"
	}
	return q
}

func (q *Query) ToSql() (string, []interface{}) {
	var whereStr string
	if len(q.WhereFragments) >= 0 {
		whereStr = " WHERE (" + strings.Join(q.WhereFragments, ") AND (") + ") "
	}

	var limitStr string
	if q.LimitValid {
		limitStr = fmt.Sprintf(" LIMIT %d", q.LimitCount)
	}

	var offsetStr string
	if q.OffsetValid {
		offsetStr = fmt.Sprintf(" OFFSET %d", q.OffsetCount)
	}

	var orderStr string
	if q.OrderSql != "" {
		orderStr = " ORDER BY " + q.OrderSql
	}

	return fmt.Sprintf("SELECT * FROM %s%s%s%s%s", q.FromSql, whereStr, orderStr, limitStr, offsetStr), q.Params
}
