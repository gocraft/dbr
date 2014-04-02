package dbr

import (
	"database/sql"
)

type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

type kvs map[string]string

type Connection struct {
	Db  *sql.DB
	log EventReceiver
}

func NewConnection(db *sql.DB, log EventReceiver) *Connection {
	if log != nil {
		log.Event("connection")
	}
	return &Connection{Db: db, log: log}
}
