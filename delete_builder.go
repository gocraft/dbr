package dbr

import (
	"database/sql"
	"fmt"
)

// DeleteBuilder builds "DELETE ..." stmt
type DeleteBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	*DeleteStmt

	LimitCount int64
}

// DeleteFrom creates a DeleteBuilder
func (sess *Session) DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		DeleteStmt:    DeleteFrom(table),
		LimitCount:    -1,
	}
}

// DeleteFrom creates a DeleteBuilder
func (tx *Tx) DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		DeleteStmt:    DeleteFrom(table),
		LimitCount:    -1,
	}
}

// DeleteBySql creates a DeleteBuilder from raw query
func (sess *Session) DeleteBySql(query string, value ...interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		DeleteStmt:    DeleteBySql(query, value...),
		LimitCount:    -1,
	}
}

// DeleteBySql creates a DeleteBuilder from raw query
func (tx *Tx) DeleteBySql(query string, value ...interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		DeleteStmt:    DeleteBySql(query, value...),
		LimitCount:    -1,
	}
}

// Exec executes the stmt
func (b *DeleteBuilder) Exec() (sql.Result, error) {
	return exec(b.runner, b.EventReceiver, b, b.Dialect)
}

// Where adds condition to the stmt
func (b *DeleteBuilder) Where(query interface{}, value ...interface{}) *DeleteBuilder {
	b.DeleteStmt.Where(query, value...)
	return b
}

// Limit adds LIMIT
func (b *DeleteBuilder) Limit(n uint64) *DeleteBuilder {
	b.LimitCount = int64(n)
	return b
}

// Build builds `DELETE ...` in dialect
func (b *DeleteBuilder) Build(d Dialect, buf Buffer) error {
	err := b.DeleteStmt.Build(b.Dialect, buf)
	if err != nil {
		return err
	}
	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(fmt.Sprint(b.LimitCount))
	}
	return nil
}
