package dbr

import (
	"database/sql"
	"fmt"
)

// UpdateBuilder builds `UPDATE ...`
type UpdateBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	*UpdateStmt

	LimitCount int64
}

// Update creates a UpdateBuilder
func (sess *Session) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		UpdateStmt:    Update(table),
		LimitCount:    -1,
	}
}

// Update creates a UpdateBuilder
func (tx *Tx) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		UpdateStmt:    Update(table),
		LimitCount:    -1,
	}
}

// UpdateBySql creates a UpdateBuilder from raw query
func (sess *Session) UpdateBySql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		UpdateStmt:    UpdateBySql(query, value...),
		LimitCount:    -1,
	}
}

// UpdateBySql creates a UpdateBuilder from raw query
func (tx *Tx) UpdateBySql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		UpdateStmt:    UpdateBySql(query, value...),
		LimitCount:    -1,
	}
}

// Exec executes the stmt
func (b *UpdateBuilder) Exec() (sql.Result, error) {
	return exec(b.runner, b.EventReceiver, b, b.Dialect)
}

// Set adds "SET column=value"
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.UpdateStmt.Set(column, value)
	return b
}

// SetMap adds "SET column=value" for each key value pair in m
func (b *UpdateBuilder) SetMap(m map[string]interface{}) *UpdateBuilder {
	b.UpdateStmt.SetMap(m)
	return b
}

// Where adds condition to the stmt
func (b *UpdateBuilder) Where(query interface{}, value ...interface{}) *UpdateBuilder {
	b.UpdateStmt.Where(query, value...)
	return b
}

// Limit adds LIMIT
func (b *UpdateBuilder) Limit(n uint64) *UpdateBuilder {
	b.LimitCount = int64(n)
	return b
}

// Build builds `UPDATE ...` in dialect
func (b *UpdateBuilder) Build(d Dialect, buf Buffer) error {
	err := b.UpdateStmt.Build(b.Dialect, buf)
	if err != nil {
		return err
	}
	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(fmt.Sprint(b.LimitCount))
	}
	return nil
}
