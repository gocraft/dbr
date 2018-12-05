// Package dbr provides additions to Go's database/sql for super fast performance and convenience.
package dbr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gocraft/dbr/dialect"
)

type ConnectionConfig struct{
	Driver string
	Dsn string
}

// Open creates a Connection.
// log can be nil to ignore logging.
func Open(driver, dsn string, log EventReceiver) (*Connection, error) {
	if log == nil {
		log = nullReceiver
	}

	// connection
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	dialect, err := loadDialect(driver)

	db := &DbAccess{DB: conn, Dialect: dialect}

	return &Connection{Read: db, Write:db, EventReceiver: log}, nil
}

// Open creates a Connection with multi read-write connections.
// log can be nil to ignore logging.
func OpenMultiConnection(readConnConfig *ConnectionConfig, writeConnConfig *ConnectionConfig, log EventReceiver) (*Connection, error) {
	if log == nil {
		log = nullReceiver
	}

	if readConnConfig == nil {
		return nil, fmt.Errorf("sql: unknown read connection configuration (forgotten?)")
	}

	if writeConnConfig == nil {
		return nil, fmt.Errorf("sql: unknown write connection configuration (forgotten?)")
	}

	// read connection
	readConn, err := sql.Open(readConnConfig.Driver, readConnConfig.Dsn)
	if err != nil {
		return nil, err
	}
	readDialect, err := loadDialect(readConnConfig.Driver)

	// write connection
	writeConn, err := sql.Open(writeConnConfig.Driver, writeConnConfig.Dsn)
	if err != nil {
		return nil, err
	}
	writeDialect, err := loadDialect(writeConnConfig.Driver)

	return &Connection{
		Read: &DbAccess{DB: readConn, Dialect: readDialect},
		Write: &DbAccess{DB: writeConn, Dialect: writeDialect},
		EventReceiver: log,
	}, nil
}

func loadDialect(driver string) (d Dialect, err error) {
	switch driver {
	case "mysql":
		d = dialect.MySQL
	case "postgres":
		d = dialect.PostgreSQL
	case "sqlite3":
		d = dialect.SQLite3
	}

	return nil, ErrNotSupported
}

const (
	placeholder = "?"
)

// Connection wraps sql.DbAccess with an EventReceiver
// to send events, errors, and timings.
type Connection struct {
	Read *DbAccess
	Write *DbAccess
	EventReceiver
}

type DbAccess struct {
	*sql.DB
	Dialect
	Timeout time.Duration
}

// Session represents a business unit of execution.
//
// All queries in gocraft/dbr are made in the context of a session.
// This is because when instrumenting your app, it's important
// to understand which business action the query took place in.
//
// A custom EventReceiver can be set.
//
// Timeout specifies max duration for an operation like Select.
type Session struct {
	*Connection
	EventReceiver

}

// GetTimeout returns current timeout enforced in session.
func (db *DbAccess) GetTimeout() time.Duration {
	return db.Timeout
}

// NewSession instantiates a Session from Connection.
// If log is nil, Connection EventReceiver is used.
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
// Both Session and Tx implements this interface.
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
	GetTimeout() time.Duration
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func exec(ctx context.Context, runner runner, log EventReceiver, builder Builder, d Dialect) (sql.Result, error) {
	timeout := runner.GetTimeout()
	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	i := interpolator{
		Buffer:       NewBuffer(),
		Dialect:      d,
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, value := i.String(), i.Value()
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

	traceImpl, hasTracingImpl := log.(TracingEventReceiver)
	if hasTracingImpl {
		ctx = traceImpl.SpanStart(ctx, "dbr.exec", query)
		defer traceImpl.SpanFinish(ctx)
	}

	result, err := runner.ExecContext(ctx, query, value...)
	if err != nil {
		if hasTracingImpl {
			traceImpl.SpanError(ctx, err)
		}
		return result, log.EventErrKv("dbr.exec.exec", err, kvs{
			"sql": query,
		})
	}
	return result, nil
}

func queryRows(ctx context.Context, runner runner, log EventReceiver, builder Builder, d Dialect) (string, *sql.Rows, error) {
	// discard the timeout set in the runner, the context should not be canceled
	// implicitly here but explicitly by the caller since the returned *sql.Rows
	// may still listening to the context
	i := interpolator{
		Buffer:       NewBuffer(),
		Dialect:      d,
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, value := i.String(), i.Value()
	if err != nil {
		return query, nil, log.EventErrKv("dbr.select.interpolate", err, kvs{
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

	traceImpl, hasTracingImpl := log.(TracingEventReceiver)
	if hasTracingImpl {
		ctx = traceImpl.SpanStart(ctx, "dbr.select", query)
		defer traceImpl.SpanFinish(ctx)
	}

	rows, err := runner.QueryContext(ctx, query, value...)
	if err != nil {
		if hasTracingImpl {
			traceImpl.SpanError(ctx, err)
		}
		return query, nil, log.EventErrKv("dbr.select.load.query", err, kvs{
			"sql": query,
		})
	}

	return query, rows, nil
}

func query(ctx context.Context, runner runner, log EventReceiver, builder Builder, d Dialect, dest interface{}) (int, error) {
	timeout := runner.GetTimeout()
	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	query, rows, err := queryRows(ctx, runner, log, builder, d)
	if err != nil {
		return 0, err
	}
	count, err := Load(rows, dest)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.load.scan", err, kvs{
			"sql": query,
		})
	}
	return count, nil
}
