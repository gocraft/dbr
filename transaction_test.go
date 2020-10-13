package dbr

import (
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestTransactionCommit(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		tx, err := sess.Begin()
		require.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		elem_count := 1
		if sess.Dialect == dialect.MSSQL {
			tx.UpdateBySql("SET IDENTITY_INSERT dbr_people ON;").Exec()
			elem_count += 1
		}

		id := 1
		result, err := tx.InsertInto("dbr_people").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Comment("INSERT TEST").Exec()
		require.NoError(t, err)
		require.Len(t, sess.EventReceiver.(*testTraceReceiver).started, elem_count)
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[elem_count-1].eventName, "dbr.exec")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[elem_count-1].query, "/* INSERT TEST */\n")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[elem_count-1].query, "INSERT")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[elem_count-1].query, "dbr_people")
		require.Contains(t, sess.EventReceiver.(*testTraceReceiver).started[elem_count-1].query, "name")
		require.Equal(t, elem_count, sess.EventReceiver.(*testTraceReceiver).finished)
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

		if sess.Dialect == dialect.MSSQL {
			tx.UpdateBySql("SET IDENTITY_INSERT dbr_people ON;").Exec()
		}

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
