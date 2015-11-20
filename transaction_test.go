package dbr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionCommit(t *testing.T) {
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		tx, err := sess.Begin()
		assert.NoError(t, err)

		result, err := tx.InsertInto("dbr_people").Columns("name").Values("A").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		err = tx.Commit()
		assert.NoError(t, err)
	}
}

func TestTransactionRollback(t *testing.T) {
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		tx, err := sess.Begin()
		assert.NoError(t, err)

		result, err := tx.InsertInto("dbr_people").Columns("name").Values("A").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		err = tx.Rollback()
		assert.NoError(t, err)
	}
}
