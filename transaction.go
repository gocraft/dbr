package dbr

import (
	"context"
	"database/sql"
	"time"
)

// Tx is a transaction created by Session.
type Tx struct {
	EventReceiver
	Dialect
	*sql.Tx
	Timeout time.Duration
}

// GetTimeout returns timeout enforced in Tx.
func (tx *Tx) GetTimeout() time.Duration {
	return tx.Timeout
}

// BeginTx creates a transaction with TxOptions.
func (sess *Session) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := sess.Connection.BeginTx(ctx, opts)
	if err != nil {
		return nil, sess.EventErr("dbr.begin.error", err)
	}
	sess.Event("dbr.begin")

	return &Tx{
		EventReceiver: sess.EventReceiver,
		Dialect:       sess.Dialect,
		Tx:            tx,
		Timeout:       sess.GetTimeout(),
	}, nil
}

// Begin creates a transaction for the given session.
func (sess *Session) Begin() (*Tx, error) {
	return sess.BeginTx(context.Background(), nil)
}

// Commit finishes the transaction.
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		return tx.EventErr("dbr.commit.error", err)
	}
	tx.Event("dbr.commit")
	return nil
}

// Rollback cancels the transaction.
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		return tx.EventErr("dbr.rollback", err)
	}
	tx.Event("dbr.rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless
// it has already been committed or rolled back.
//
// Useful to defer tx.RollbackUnlessCommitted(), so you don't
// have to handle N failure cases.
// Keep in mind the only way to detect an error on the rollback
// is via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	err := tx.Tx.Rollback()
	if err == sql.ErrTxDone {
		// ok
	} else if err != nil {
		tx.EventErr("dbr.rollback_unless_committed", err)
	} else {
		tx.Event("dbr.rollback")
	}
}
