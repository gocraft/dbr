package dbr

import (
	"database/sql"
)

type Eq map[string]interface{}

type Connection struct {
	Db *sql.DB
	EventReceiver
}

type Session struct {
	cxn *Connection
	EventReceiver
}

type SessionRunner interface {
	Select(cols ...string) *SelectBuilder
	SelectBySql(sql string, args ...interface{}) *SelectBuilder
	
	InsertInto(into string) *InsertBuilder
	Update(table string) *UpdateBuilder
	DeleteFrom(from string) *DeleteBuilder
}

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func NewConnection(db *sql.DB, log EventReceiver) *Connection {
	if log == nil {
		log = nullReceiver
	}

	return &Connection{Db: db, EventReceiver: log}
}

func (cxn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = cxn.EventReceiver // Use parent instrumentation
	}
	return &Session{cxn: cxn, EventReceiver: log}
}
