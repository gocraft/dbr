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

func (cxn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = cxn.EventReceiver // Use parent instrumentation
	}
	return &Session{cxn: cxn, EventReceiver: log}
}

func NewConnection(db *sql.DB, log EventReceiver) *Connection {
	if log == nil {
		log = nullReceiver
	}

	return &Connection{Db: db, EventReceiver: log}
}
