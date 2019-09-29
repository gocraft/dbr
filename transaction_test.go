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

		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Comment("INSERT TEST").Exec()
		require.NoError(t, err)
		require.Len(t, sess.EventReceiver.(*testTraceReceiver).started, 1)
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[0].eventName, "dbr.exec")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[0].query, "/* INSERT TEST */\n")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[0].query, "INSERT")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[0].query, "dbr_people")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[0].query, "name")
		require.Equal(t, 1, sess.EventReceiver.(*testTraceReceiver).finished)
		require.Equal(t, 0, sess.EventReceiver.(*testTraceReceiver).errored)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		err = tx.Commit()
		require.NoError(t, err)

		var person dbrPerson
		err = tx.Select("*").From("dbr_people").Where(Eq("id", id)).LoadOne(&person)
		require.Error(t, err)
		require.Equal(t, 1, sess.EventReceiver.(*testTraceReceiver).errored)
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
