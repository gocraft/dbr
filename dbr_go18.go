// +build go1.8

package dbr

import (
	"database/sql"
)

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (sess *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return sess.ExecContext(sess.ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (sess *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return sess.QueryContext(sess.ctx, query, args...)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(tx.ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(tx.ctx, query, args...)
}

// beginTx starts a transaction with context.
func (sess *Session) beginTx() (*sql.Tx, error) {
	return sess.BeginTx(sess.ctx, nil)
}
