package dbr

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEventReceiverWithContextSessionLoad(t *testing.T) {
	sess := postgresSession
	reset(t, sess)

	ctx := context.WithValue(context.Background(), "key", "value")
	var person dbrPerson
	err := sess.InsertInto("dbr_people").Columns("name").Record(&person).
		Returning("id").LoadContext(ctx, &person.Id)
	require.NoError(t, err)
	require.Equal(t, ctx, sess.EventReceiverWithContext.(*testTraceReceiverWithContext).ctx)
}

func TestEventReceiverWithContextSessionExec(t *testing.T) {
	sess := postgresSession
	reset(t, sess)

	ctx := context.WithValue(context.Background(), "key", "value")
	var person dbrPerson
	_, err := sess.InsertInto("dbr_people").Columns("name").Record(&person).
		ExecContext(ctx)
	require.NoError(t, err)
	require.Equal(t, ctx, sess.EventReceiverWithContext.(*testTraceReceiverWithContext).ctx)
}

func TestEventReceiverWithContextTxLoad(t *testing.T) {
	sess := postgresSession
	reset(t, sess)

	ctx := context.WithValue(context.Background(), "key", "value")
	var person dbrPerson
	tx, _ := sess.BeginTx(ctx, nil)
	err := tx.InsertInto("dbr_people").Columns("name").Record(&person).
		Returning("id").LoadContext(ctx, &person.Id)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)
	require.Equal(t, ctx, sess.EventReceiverWithContext.(*testTraceReceiverWithContext).ctx)
}

func TestEventReceiverWithContextTxExec(t *testing.T) {
	sess := postgresSession
	reset(t, sess)

	ctx := context.WithValue(context.Background(), "key", "value")
	var person dbrPerson
	tx, _ := sess.BeginTx(ctx, nil)
	_, err := tx.InsertInto("dbr_people").Columns("name").Record(&person).
		ExecContext(ctx)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)
	require.Equal(t, ctx, sess.EventReceiverWithContext.(*testTraceReceiverWithContext).ctx)
}
