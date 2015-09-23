package dbr

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gocraft/dbr/dialect"
)

// Open instantiates a Connection for a given database/sql connection
// and event receiver
func Open(driver, dsn string, log EventReceiver) (*Connection, error) {
	if log == nil {
		log = nullReceiver
	}
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	var d Dialect
	switch driver {
	case "mysql":
		d = dialect.MySQL
	case "postgres":
		d = dialect.PostgreSQL
	default:
		return nil, ErrNotSupported
	}
	return &Connection{DB: conn, EventReceiver: log, Dialect: d}, nil
}

// Don't break the API
// FIXME: This will be removed in the future
func NewConnection(db *sql.DB, log EventReceiver) *Connection {
	if log == nil {
		log = nullReceiver
	}
	return &Connection{DB: db, EventReceiver: log, Dialect: dialect.MySQL}
}

// Connection is a connection to the database with an EventReceiver
// to send events, errors, and timings to
type Connection struct {
	*sql.DB
	Dialect Dialect
	EventReceiver
}

// Session represents a business unit of execution for some connection
type Session struct {
	*Connection
	EventReceiver
}

// NewSession instantiates a Session for the Connection
func (conn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = conn.EventReceiver // Use parent instrumentation
	}
	return &Session{Connection: conn, EventReceiver: log}
}

// Ensure that tx and session are session runner
var (
	_ SessionRunner = (*Tx)(nil)
	_ SessionRunner = (*Session)(nil)
)

// SessionRunner can do anything that a Session can except start a transaction.
type SessionRunner interface {
	Select(column ...string) *SelectBuilder
	SelectBySql(query string, value ...interface{}) *SelectBuilder

	InsertInto(table string) *InsertBuilder
	InsertBySql(query string, value ...interface{}) *InsertBuilder

	Update(table string) *UpdateBuilder
	UpdateBySql(query string, value ...interface{}) *UpdateBuilder

	DeleteFrom(table string) *DeleteBuilder
	DeleteBySql(query string, value ...interface{}) *DeleteBuilder
}

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type builder interface {
	ToSql() (string, []interface{})
}

func exec(runner runner, log EventReceiver, builder builder, d Dialect) (sql.Result, error) {
	query, value := builder.ToSql()
	query, err := InterpolateForDialect(query, value, d)
	if err != nil {
		return nil, log.EventErrKv("dbr.exec.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("dbr.exec", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	result, err := runner.Exec(query)
	if err != nil {
		return result, log.EventErrKv("dbr.exec.exec", err, kvs{
			"sql": query,
		})
	}
	return result, nil
}

func query(runner runner, log EventReceiver, builder builder, d Dialect, v interface{}) (int, error) {
	query, value := builder.ToSql()
	query, err := InterpolateForDialect(query, value, d)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	rows, err := runner.Query(query)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.load.query", err, kvs{
			"sql": query,
		})
	}
	count, err := Load(rows, v)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.load.scan", err, kvs{
			"sql": query,
		})
	}
	return count, nil
}
