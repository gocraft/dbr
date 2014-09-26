package dbr

import (
	"database/sql"
)

// Connection is a connection to the database with an EventReceiver
// to send events, errors, and timings to
type Connection struct {
	Db *sql.DB
	EventReceiver
}

// Session represents a business unit of execution for some connection
type Session struct {
	cxn *Connection
	EventReceiver
}

// NewConnection instantiates a Connection for a given database/sql connection
// and event receiver
func NewConnection(db *sql.DB, log EventReceiver) *Connection {
	if log == nil {
		log = nullReceiver
	}

	return &Connection{Db: db, EventReceiver: log}
}

// NewSession instantiates a Session for the Connection
func (cxn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = cxn.EventReceiver // Use parent instrumentation
	}
	return &Session{cxn: cxn, EventReceiver: log}
}

// SessionRunner can do anything that a Session can except start a transaction.
type SessionRunner interface {
	Select(cols ...string) *SelectBuilder
	SelectBySql(sql string, args ...interface{}) *SelectBuilder

	InsertInto(into string) *InsertBuilder
	Update(table string) *UpdateBuilder
	UpdateBySql(sql string, args ...interface{}) *UpdateBuilder
	DeleteFrom(from string) *DeleteBuilder
}

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
