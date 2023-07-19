package dbr

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/gocraft/dbr/v2/dialect"
)

// SelectStmt builds `SELECT ...`.
type SelectStmt struct {
	Runner
	EventReceiver
	Dialect

	raw

	IsDistinct bool

	Column    []interface{}
	Table     interface{}
	JoinTable []Builder

	WhereCond  []Builder
	Group      []Builder
	HavingCond []Builder
	Order      []Builder
	Suffixes   []Builder

	LimitCount  int64
	OffsetCount int64

	comments Comments

	indexHints []Builder
}

type SelectBuilder = SelectStmt

func (b *SelectStmt) Build(d Dialect, buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(d, buf)
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	err := b.comments.Build(d, buf)
	if err != nil {
		return err
	}

	buf.WriteString("SELECT ")

	if b.IsDistinct {
		buf.WriteString("DISTINCT ")
	}

	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(", ")
		}
		switch col := col.(type) {
		case string:
			// FIXME: no quote ident
			buf.WriteString(col)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(col)
		}
	}

	if b.Table != nil {
		buf.WriteString(" FROM ")
		switch table := b.Table.(type) {
		case string:
			// FIXME: no quote ident
			buf.WriteString(table)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(table)
		}

		for _, hint := range b.indexHints {
			buf.WriteString(" ")
			if err := hint.Build(d, buf); err != nil {
				return err
			}
		}

		if len(b.JoinTable) > 0 {
			for _, join := range b.JoinTable {
				err := join.Build(d, buf)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}

	if len(b.Group) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, group := range b.Group {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := group.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	if len(b.HavingCond) > 0 {
		buf.WriteString(" HAVING ")
		err := And(b.HavingCond...).Build(d, buf)
		if err != nil {
			return err
		}
	}

	if len(b.Order) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, order := range b.Order {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := order.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	if d == dialect.MSSQL {
		b.addMSSQLLimits(buf)
	} else {
		if b.LimitCount >= 0 {
			buf.WriteString(" LIMIT ")
			buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
		}

		if b.OffsetCount >= 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.FormatInt(b.OffsetCount, 10))
		}
	}

	if len(b.Suffixes) > 0 {
		for _, suffix := range b.Suffixes {
			buf.WriteString(" ")
			err := suffix.Build(d, buf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// https://docs.microsoft.com/en-us/previous-versions/sql/sql-server-2012/ms188385(v=sql.110)
func (b *SelectStmt) addMSSQLLimits(buf Buffer) {
	limitCount := b.LimitCount
	offsetCount := b.OffsetCount
	if limitCount < 0 && offsetCount < 0 {
		return
	}
	if offsetCount < 0 {
		offsetCount = 0
	}

	if len(b.Order) == 0 {
		// ORDER is required for OFFSET / FETCH
		buf.WriteString(" ORDER BY ")
		col := b.Column[0]
		switch col := col.(type) {
		case string:
			// FIXME: no quote ident
			buf.WriteString(col)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(col)
		}
	}

	buf.WriteString(" OFFSET ")
	buf.WriteString(strconv.FormatInt(offsetCount, 10))
	buf.WriteString(" ROWS ")

	if limitCount >= 0 {
		buf.WriteString(" FETCH FIRST ")
		buf.WriteString(strconv.FormatInt(limitCount, 10))
		buf.WriteString(" ROWS ONLY ")
	}
}

// Select creates a SelectStmt.
func Select(column ...interface{}) *SelectStmt {
	return &SelectStmt{
		Column:      column,
		LimitCount:  -1,
		OffsetCount: -1,
	}
}

func prepareSelect(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

// Select creates a SelectStmt.
func (sess *Session) Select(column ...string) *SelectStmt {
	b := Select(prepareSelect(column)...)
	b.Runner = sess
	b.EventReceiver = sess.EventReceiver
	b.Dialect = sess.Dialect
	return b
}

// Select creates a SelectStmt.
func (tx *Tx) Select(column ...string) *SelectStmt {
	b := Select(prepareSelect(column)...)
	b.Runner = tx
	b.EventReceiver = tx.EventReceiver
	b.Dialect = tx.Dialect
	return b
}

// SelectBySql creates a SelectStmt from raw query.
func SelectBySql(query string, value ...interface{}) *SelectStmt {
	return &SelectStmt{
		raw: raw{
			Query: query,
			Value: value,
		},
		LimitCount:  -1,
		OffsetCount: -1,
	}
}

// SelectBySql creates a SelectStmt from raw query.
func (sess *Session) SelectBySql(query string, value ...interface{}) *SelectStmt {
	b := SelectBySql(query, value...)
	b.Runner = sess
	b.EventReceiver = sess.EventReceiver
	b.Dialect = sess.Dialect
	return b
}

// SelectBySql creates a SelectStmt from raw query.
func (tx *Tx) SelectBySql(query string, value ...interface{}) *SelectStmt {
	b := SelectBySql(query, value...)
	b.Runner = tx
	b.EventReceiver = tx.EventReceiver
	b.Dialect = tx.Dialect
	return b
}

// From specifies table to select from.
// table can be Builder like SelectStmt, or string.
func (b *SelectStmt) From(table interface{}) *SelectStmt {
	b.Table = table
	return b
}

func (b *SelectStmt) Distinct() *SelectStmt {
	b.IsDistinct = true
	return b
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectStmt) Where(query interface{}, value ...interface{}) *SelectStmt {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Having adds a having condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectStmt) Having(query interface{}, value ...interface{}) *SelectStmt {
	switch query := query.(type) {
	case string:
		b.HavingCond = append(b.HavingCond, Expr(query, value...))
	case Builder:
		b.HavingCond = append(b.HavingCond, query)
	}
	return b
}

// GroupBy specifies columns for grouping.
func (b *SelectStmt) GroupBy(col ...string) *SelectStmt {
	for _, group := range col {
		b.Group = append(b.Group, Expr(group))
	}
	return b
}

func (b *SelectStmt) OrderAsc(col string) *SelectStmt {
	b.Order = append(b.Order, order(col, asc))
	return b
}

func (b *SelectStmt) OrderDesc(col string) *SelectStmt {
	b.Order = append(b.Order, order(col, desc))
	return b
}

// OrderBy specifies columns for ordering.
func (b *SelectStmt) OrderBy(col string) *SelectStmt {
	b.Order = append(b.Order, Expr(col))
	return b
}

func (b *SelectStmt) Limit(n uint64) *SelectStmt {
	b.LimitCount = int64(n)
	return b
}

func (b *SelectStmt) Offset(n uint64) *SelectStmt {
	b.OffsetCount = int64(n)
	return b
}

// Suffix adds an expression to the end of the query. This is useful to add dialect-specific clauses like FOR UPDATE
func (b *SelectStmt) Suffix(suffix string, value ...interface{}) *SelectStmt {
	b.Suffixes = append(b.Suffixes, Expr(suffix, value...))
	return b
}

// Paginate fetches a page in a naive way for a small set of data.
func (b *SelectStmt) Paginate(page, perPage uint64) *SelectStmt {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// OrderDir is a helper for OrderAsc and OrderDesc.
func (b *SelectStmt) OrderDir(col string, isAsc bool) *SelectStmt {
	if isAsc {
		b.OrderAsc(col)
	} else {
		b.OrderDesc(col)
	}
	return b
}

func (b *SelectStmt) Comment(comment string) *SelectStmt {
	b.comments = b.comments.Append(comment)
	return b
}

// Join add inner-join.
// on can be Builder or string.
func (b *SelectStmt) Join(table, on interface{}, indexHints ...Builder) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(inner, table, on, indexHints))
	return b
}

// LeftJoin add left-join.
// on can be Builder or string.
func (b *SelectStmt) LeftJoin(table, on interface{}, indexHints ...Builder) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(left, table, on, indexHints))
	return b
}

// RightJoin add right-join.
// on can be Builder or string.
func (b *SelectStmt) RightJoin(table, on interface{}, indexHints ...Builder) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(right, table, on, indexHints))
	return b
}

// FullJoin add full-join.
// on can be Builder or string.
func (b *SelectStmt) FullJoin(table, on interface{}, indexHints ...Builder) *SelectStmt {
	b.JoinTable = append(b.JoinTable, join(full, table, on, indexHints))
	return b
}

// As creates alias for select statement.
func (b *SelectStmt) As(alias string) Builder {
	return as(b, alias)
}

// Rows executes the query and returns the rows returned, or any error encountered.
func (b *SelectStmt) Rows() (*sql.Rows, error) {
	return b.RowsContext(context.Background())
}

func (b *SelectStmt) RowsContext(ctx context.Context) (*sql.Rows, error) {
	_, rows, err := queryRows(ctx, b.Runner, b.EventReceiver, b, b.Dialect)
	return rows, err
}

func (b *SelectStmt) LoadOneContext(ctx context.Context, value interface{}) error {
	count, err := query(ctx, b.Runner, b.EventReceiver, b, b.Dialect, value)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// LoadOne loads SQL result into go variable that is not a slice.
// Unlike Load, it returns ErrNotFound if the SQL result row count is 0.
//
// See https://godoc.org/github.com/gocraft/dbr#Load.
func (b *SelectStmt) LoadOne(value interface{}) error {
	return b.LoadOneContext(context.Background(), value)
}

func (b *SelectStmt) LoadContext(ctx context.Context, value interface{}) (int, error) {
	return query(ctx, b.Runner, b.EventReceiver, b, b.Dialect, value)
}

// Load loads multi-row SQL result into a slice of go variables.
//
// See https://godoc.org/github.com/gocraft/dbr#Load.
func (b *SelectStmt) Load(value interface{}) (int, error) {
	return b.LoadContext(context.Background(), value)
}

// Iterate executes the query and returns the Iterator, or any error encountered.
func (b *SelectStmt) Iterate() (Iterator, error) {
	return b.IterateContext(context.Background())
}

// IterateContext executes the query and returns the Iterator, or any error encountered.
func (b *SelectStmt) IterateContext(ctx context.Context) (Iterator, error) {
	_, rows, err := queryRows(ctx, b.Runner, b.EventReceiver, b, b.Dialect)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}
	iterator := iteratorInternals{
		rows:    rows,
		columns: columns,
	}
	return &iterator, err
}

// IndexHint adds a index hint.
// hint can be Builder or string.
func (b *SelectStmt) IndexHint(hints ...interface{}) *SelectStmt {
	for _, hint := range hints {
		switch hint := hint.(type) {
		case string:
			b.indexHints = append(b.indexHints, Expr(hint))
		case Builder:
			b.indexHints = append(b.indexHints, hint)
		}
	}
	return b
}
