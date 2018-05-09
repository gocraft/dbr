package dbr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionCommit(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		tx, err := sess.Begin()
		require.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := 1

		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		err = tx.Commit()
		require.NoError(t, err)

		var person dbrPerson
		err = tx.Select("*").From("dbr_people").Where(Eq("id", id)).LoadOne(&person)
		require.Error(t, err)
	}
}

func TestTransactionRollback(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		tx, err := sess.Begin()
		require.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := 1

		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		err = tx.Rollback()
		require.NoError(t, err)

		var person dbrPerson
		err = tx.Select("*").From("dbr_people").Where(Eq("id", id)).LoadOne(&person)
		require.Error(t, err)
	}
}
