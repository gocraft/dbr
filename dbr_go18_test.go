// +build go1.8

package dbr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextCancel(t *testing.T) {
	// context support is implemented for PostgreSQL
	checkSessionContext(t, postgresSession.Connection)
	checkTxQueryContext(t, postgresSession.Connection)
	checkTxExecContext(t, postgresSession.Connection)
}

func checkSessionContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sess := conn.NewSessionContext(ctx, nil)
	_, err := sess.SelectBySql("SELECT 1").ReturnInt64()
	assert.EqualError(t, err, "context canceled")
	_, err = sess.Update("dbr_people").Where(Eq("id", 1)).Set("name", "jonathan1").Exec()
	assert.EqualError(t, err, "context canceled")
	_, err = sess.Begin()
	assert.EqualError(t, err, "context canceled")
}

func checkTxQueryContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	sess := conn.NewSessionContext(ctx, nil)
	tx, err := sess.Begin()
	if !assert.NoError(t, err) {
		cancel()
		return
	}
	cancel()
	_, err = tx.SelectBySql("SELECT 1").ReturnInt64()
	assert.EqualError(t, err, "context canceled")
	assert.NoError(t, tx.Rollback())
}

func checkTxExecContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	sess := conn.NewSessionContext(ctx, nil)
	tx, err := sess.Begin()
	if !assert.NoError(t, err) {
		cancel()
		return
	}
	_, err = tx.Update("dbr_people").Where(Eq("id", 1)).Set("name", "jonathan1").Exec()
	assert.NoError(t, err)
	cancel()
	assert.EqualError(t, tx.Commit(), "context canceled")
}
