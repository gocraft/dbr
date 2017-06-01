package dbr

// SelectBuilder build "SELECT" stmt
type SelectBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	*SelectStmt
}

func prepareSelect(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

// Select creates a SelectBuilder
func (sess *Session) Select(column ...string) *SelectBuilder {
	return &SelectBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		SelectStmt:    Select(prepareSelect(column)...),
	}
}

// Select creates a SelectBuilder
func (tx *Tx) Select(column ...string) *SelectBuilder {
	return &SelectBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		SelectStmt:    Select(prepareSelect(column)...),
	}
}

// SelectBySql creates a SelectBuilder from raw query
func (sess *Session) SelectBySql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		SelectStmt:    SelectBySql(query, value...),
	}
}

// SelectBySql creates a SelectBuilder from raw query
func (tx *Tx) SelectBySql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		SelectStmt:    SelectBySql(query, value...),
	}
}

// Load loads any value from query result
func (b *SelectBuilder) Load(value interface{}) (int, error) {
	return query(b.runner, b.EventReceiver, b, b.Dialect, value)
}

// LoadStruct loads struct from query result, returns ErrNotFound if there is no result
func (b *SelectBuilder) LoadStruct(value interface{}) error {
	count, err := query(b.runner, b.EventReceiver, b, b.Dialect, value)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// LoadStructs loads structures from query result
func (b *SelectBuilder) LoadStructs(value interface{}) (int, error) {
	return query(b.runner, b.EventReceiver, b, b.Dialect, value)
}

// LoadValue loads any value from query result, returns ErrNotFound if there is no result
func (b *SelectBuilder) LoadValue(value interface{}) error {
	count, err := query(b.runner, b.EventReceiver, b, b.Dialect, value)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// LoadValues loads any values from query result
func (b *SelectBuilder) LoadValues(value interface{}) (int, error) {
	return query(b.runner, b.EventReceiver, b, b.Dialect, value)
}

// Join joins table on condition
func (b *SelectBuilder) Join(table, on interface{}) *SelectBuilder {
	b.SelectStmt.Join(table, on)
	return b
}

// LeftJoin joins table on condition via LEFT JOIN
func (b *SelectBuilder) LeftJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.LeftJoin(table, on)
	return b
}

// RightJoin joins table on condition via RIGHT JOIN
func (b *SelectBuilder) RightJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.RightJoin(table, on)
	return b
}

// FullJoin joins table on condition via FULL JOIN
func (b *SelectBuilder) FullJoin(table, on interface{}) *SelectBuilder {
	b.SelectStmt.FullJoin(table, on)
	return b
}

// Distinct adds `DISTINCT`
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.SelectStmt.Distinct()
	return b
}

// From specifies table
func (b *SelectBuilder) From(table interface{}) *SelectBuilder {
	b.SelectStmt.From(table)
	return b
}

// GroupBy specifies columns for grouping
func (b *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	b.SelectStmt.GroupBy(col...)
	return b
}

// Having adds a having condition
func (b *SelectBuilder) Having(query interface{}, value ...interface{}) *SelectBuilder {
	b.SelectStmt.Having(query, value...)
	return b
}

// Limit adds LIMIT
func (b *SelectBuilder) Limit(n uint64) *SelectBuilder {
	b.SelectStmt.Limit(n)
	return b
}

// Offset adds OFFSET
func (b *SelectBuilder) Offset(n uint64) *SelectBuilder {
	b.SelectStmt.Offset(n)
	return b
}

// OrderDir specifies columns for ordering in direction
func (b *SelectBuilder) OrderDir(col string, isAsc bool) *SelectBuilder {
	if isAsc {
		b.SelectStmt.OrderAsc(col)
	} else {
		b.SelectStmt.OrderDesc(col)
	}
	return b
}

// Paginate adds LIMIT and OFFSET
func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// OrderBy specifies column for ordering
func (b *SelectBuilder) OrderBy(col string) *SelectBuilder {
	b.SelectStmt.Order = append(b.SelectStmt.Order, Expr(col))
	return b
}

// Where adds a where condition
func (b *SelectBuilder) Where(query interface{}, value ...interface{}) *SelectBuilder {
	b.SelectStmt.Where(query, value...)
	return b
}
