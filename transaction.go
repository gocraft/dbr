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

func (tx *Tx) Commit() {

}
