package dbr

import "database/sql"

// Tx is a transaction for the given Session
type Tx struct {
	EventReceiver
	Dialect Dialect
	*sql.Tx
}

// Begin creates a transaction for the given session
func (sess *Session) Begin() (*Tx, error) {
	tx, err := sess.Connection.Begin()
	if err != nil {
		return nil, sess.EventErr("dbr.begin.error", err)
	}
	sess.Event("dbr.begin")

	return &Tx{
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		Tx:            tx,
	}, nil
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		return tx.EventErr("dbr.commit.error", err)
	}
	tx.Event("dbr.commit")
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		return tx.EventErr("dbr.rollback", err)
	}
	tx.Event("dbr.rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless it has already been committed or rolled back.
// Useful to defer tx.RollbackUnlessCommitted() -- so you don't have to handle N failure cases
// Keep in mind the only way to detect an error on the rollback is via the event log.
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
