package dbr

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	conn := &Connection{
		DB:            db,
		EventReceiver: &NullEventReceiver{},
		Dialect:       dialect.MySQL,
	}
	sess := conn.NewSession(nil)

	mock.ExpectQuery("SELECT id FROM suggestions").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	id, err := sess.Select("id").From("suggestions").ReturnInt64s()
	require.NoError(t, err)
	require.Equal(t, []int64{1, 2}, id)

	mock.ExpectClose()
	conn.Close()

	require.NoError(t, mock.ExpectationsWereMet())
}
