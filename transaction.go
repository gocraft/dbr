package dbr

import (
	"database/sql"
)

type Tx struct {
	*Session
	*sql.Tx
}

func (sess *Session) Begin() (*Tx, error) {
	tx, err := sess.cxn.Db.Begin()
	if err != nil {
		return nil, sess.EventErr("dbr.begin", err)
	}

	return &Tx{
		Session: sess,
		Tx:      tx,
	}, nil
}

func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		return tx.EventErr("dbr.commit", err)
	}
	return nil
}

func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		return tx.EventErr("dbr.rollback", err)
	}
	return nil
}
