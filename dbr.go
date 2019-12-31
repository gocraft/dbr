// Package dbr provides additions to Go's database/sql for super fast performance and convenience.
package dbr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gocraft/dbr/v2/dialect"
)

// Open creates a Connection.
// log can be nil to ignore logging.
func Open(driver, dsn string, log EventReceiver) (*Connection, error) {
	return open(driver, dsn, log, nil)
}

func open(driver, dsn string, log EventReceiver, logWithContext EventReceiverWithContext) (*Connection, error) {
	if log == nil {
		log = nullReceiver
	}
	if logWithContext == nil {
		logWithContext = nullReceiverWithContext
	}
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	var d Dialect
	switch driver {
	case "mysql":
		d = dialect.MySQL
	case "postgres", "pgx":
		d = dialect.PostgreSQL
	case "sqlite3":
		d = dialect.SQLite3
	default:
		return nil, ErrNotSupported
	}
	return &Connection{DB: conn, EventReceiver: log, EventReceiverWithContext: logWithContext, Dialect: d}, nil

}

// Open creates a Connection.
// log implements EventReceiverWithContext
// (can be nil to ignore logging).
func OpenForLogWithContext(driver, dsn string, log EventReceiverWithContext) (*Connection, error) {
	return open(driver, dsn, nil, log)
}

const (
	placeholder = "?"
)

// Connection wraps sql.DB with an EventReceiver
// to send events, errors, and timings.
type Connection struct {
	*sql.DB
	Dialect
	EventReceiver
	EventReceiverWithContext
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
	EventReceiverWithContext
	Timeout time.Duration
}

// GetTimeout returns current timeout enforced in session.
func (sess *Session) GetTimeout() time.Duration {
	return sess.Timeout
}

// NewSession instantiates a Session from Connection.
// If log is nil, Connection EventReceiver is used.
func (conn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = conn.EventReceiver // Use parent instrumentation
	}
	return &Session{Connection: conn, EventReceiver: log, EventReceiverWithContext: conn.EventReceiverWithContext}
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

func exec(ctx context.Context, runner runner, log EventReceiver, logWithContext EventReceiverWithContext, builder Builder, d Dialect) (sql.Result, error) {
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
		eventKvs := kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		}
		eventErrKv := logWithContext.EventErrKvWithContext(ctx, "dbr.exec.interpolate", err, eventKvs)
		return nil, log.EventErrKv("dbr.exec.interpolate", eventErrKv, eventKvs)
	}

	startTime := time.Now()
	defer func() {
		timing := time.Since(startTime).Nanoseconds()
		timingKvs := kvs{
			"sql": query,
		}
		logWithContext.TimingKvWithContext(ctx, "dbr.exec", timing, timingKvs)
		log.TimingKv("dbr.exec", timing, timingKvs)
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
		eventKvs := kvs{
			"sql": query,
		}
		eventErrKv := logWithContext.EventErrKvWithContext(ctx, "dbr.exec.exec", err, eventKvs)
		return result, log.EventErrKv("dbr.exec.exec", eventErrKv, eventKvs)
	}
	return result, nil
}

func queryRows(ctx context.Context, runner runner, log EventReceiver, logWithContext EventReceiverWithContext, builder Builder, d Dialect) (string, *sql.Rows, error) {
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
		eventKvs := kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		}
		eventErrKv := logWithContext.EventErrKvWithContext(ctx, "dbr.select.interpolate", err, eventKvs)
		return query, nil, log.EventErrKv("dbr.select.interpolate", eventErrKv, eventKvs)
	}

	startTime := time.Now()
	defer func() {
		timing := time.Since(startTime).Nanoseconds()
		timingKvs := kvs{
			"sql": query,
		}
		logWithContext.TimingKvWithContext(ctx, "dbr.select", timing, timingKvs)
		log.TimingKv("dbr.select", timing, timingKvs)
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
		eventKvs := kvs{
			"sql": query,
		}
		eventErrKv := logWithContext.EventErrKvWithContext(ctx, "dbr.select.load.query", err, eventKvs)
		return query, nil, log.EventErrKv("dbr.select.load.query", eventErrKv, eventKvs)
	}

	return query, rows, nil
}

func query(ctx context.Context, runner runner, log EventReceiver, logWithContext EventReceiverWithContext, builder Builder, d Dialect, dest interface{}) (int, error) {
	timeout := runner.GetTimeout()
	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	query, rows, err := queryRows(ctx, runner, log, logWithContext, builder, d)
	if err != nil {
		return 0, err
	}
	count, err := Load(rows, dest)
	if err != nil {
		eventKvs := kvs{
			"sql": query,
		}
		eventErrKv := logWithContext.EventErrKvWithContext(ctx, "dbr.select.load.scan", err, eventKvs)
		return 0, log.EventErrKv("dbr.select.load.scan", eventErrKv, eventKvs)
	}
	return count, nil
}
