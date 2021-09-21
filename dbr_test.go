package dbr

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr/v2/dialect"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

//
// Test helpers
//

var (
	mysqlDSN    = os.Getenv("DBR_TEST_MYSQL_DSN")
	postgresDSN = os.Getenv("DBR_TEST_POSTGRES_DSN")
	sqlite3DSN  = ":memory:"
	mssqlDSN    = os.Getenv("DBR_TEST_MSSQL_DSN")
)

func createSession(driver, dsn string) *Session {
	conn, err := Open(driver, dsn, &testTraceReceiver{})
	if err != nil {
		panic(err)
	}
	return conn.NewSession(nil)
}

var (
	mysqlSession          = createSession("mysql", mysqlDSN)
	postgresSession       = createSession("postgres", postgresDSN)
	postgresBinarySession = createSession("postgres", postgresDSN+"&binary_parameters=yes")
	sqlite3Session        = createSession("sqlite3", sqlite3DSN)
	mssqlSession          = createSession("mssql", mssqlDSN)

	// all test sessions should be here
	testSession = []*Session{mysqlSession, postgresSession, sqlite3Session, mssqlSession}
)

type dbrPerson struct {
	Id    int64
	Name  string
	Email string
}

type nullTypedRecord struct {
	Id         int64
	StringVal  NullString
	Int64Val   NullInt64
	Float64Val NullFloat64
	TimeVal    NullTime
	BoolVal    NullBool
}

func reset(t *testing.T, sess *Session) {
	autoIncrementType := "serial PRIMARY KEY"
	boolType := "bool"
	datetimeType := "timestamp"

	switch sess.Dialect {
	case dialect.SQLite3:
		autoIncrementType = "integer PRIMARY KEY"
	case dialect.MSSQL:
		autoIncrementType = "integer IDENTITY PRIMARY KEY"
		boolType = "BIT"
		datetimeType = "datetime"
	}
	for _, v := range []string{
		`DROP TABLE IF EXISTS dbr_people`,
		fmt.Sprintf(`CREATE TABLE dbr_people (
			id %s,
			name varchar(255) NOT NULL,
			email varchar(255)
		)`, autoIncrementType),

		`DROP TABLE IF EXISTS dbr_friend`,
		fmt.Sprintf(`CREATE TABLE dbr_friend (
			one integer NOT NULL,
			another integer NOT NULL
		)`),

		`DROP TABLE IF EXISTS null_types`,
		fmt.Sprintf(`CREATE TABLE null_types (
			id %s,
			string_val varchar(255) NULL,
			int64_val integer NULL,
			float64_val float NULL,
			time_val %s NULL,
			bool_val %s NULL
		)`, autoIncrementType, datetimeType, boolType),
	} {
		_, err := sess.Exec(v)
		require.NoError(t, err)
	}
	// clear test data collected by testTraceReceiver
	sess.EventReceiver = &testTraceReceiver{}
}

func TestBasicCRUD(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		jonathan := dbrPerson{
			Name:  "jonathan",
			Email: "jonathan@uservoice.com",
		}
		bamiyan := dbrPerson{
			Name:  "bamiyan",
			Email: "bamiyan@uservoice.com",
		}
		insertColumns := []string{"name", "email"}
		if sess.Dialect == dialect.PostgreSQL {
			jonathan.Id = 1
			bamiyan.Id = 2
			insertColumns = []string{"id", "name", "email"}
		}
		if sess.Dialect == dialect.MSSQL {
			jonathan.Id = 1
			bamiyan.Id = 2
		}

		// insert
		result, err := sess.InsertInto("dbr_people").Columns(insertColumns...).Record(&jonathan).Exec()
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		require.True(t, jonathan.Id > 0)

		result, err = sess.InsertInto("dbr_people").Columns(insertColumns...).Record(&bamiyan).Exec()
		require.NoError(t, err)

		result, err = sess.InsertInto("dbr_friend").Pair("one", jonathan.Id).Pair("another", bamiyan.Id).Exec()
		require.NoError(t, err)

		// select
		var people []dbrPerson
		count, err := sess.Select("*").From("dbr_people").Where(Eq("id", jonathan.Id)).Load(&people)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, jonathan.Id, people[0].Id)
		require.Equal(t, jonathan.Name, people[0].Name)
		require.Equal(t, jonathan.Email, people[0].Email)

		// select id
		ids, err := sess.Select("id").From("dbr_people").ReturnInt64s()
		require.NoError(t, err)
		require.Equal(t, 2, len(ids))

		// select id limit
		ids, err = sess.Select("id").From("dbr_people").Limit(1).ReturnInt64s()
		require.NoError(t, err)
		require.Equal(t, 1, len(ids))

		friends := struct {
			One struct {
				Id    int64  `db:"id"`
				Name  string `db:"name"`
				Email string `db:"email"`
			} `db:"one"`
			Another struct {
				Id    int64  `db:"id"`
				Name  string `db:"name"`
				Email string `db:"email"`
			} `db:"another"`
		}{}

		err = sess.Select("one.id AS \"one.id\"", "one.name AS \"one.name\"", "one.email AS \"one.email\"", "another.id AS \"another.id\"", "another.name AS \"another.name\"", "another.email AS \"another.email\"").
			From(I("dbr_people").As("one")).
			Join("dbr_friend", "one.id = dbr_friend.one").
			Join(I("dbr_people").As("another"), "another.id = dbr_friend.another").
			LoadOne(&friends)
		require.NoError(t, err)
		require.Equal(t, jonathan.Id, friends.One.Id)
		require.Equal(t, jonathan.Name, friends.One.Name)
		require.Equal(t, jonathan.Name, friends.One.Name)
		require.Equal(t, bamiyan.Id, friends.Another.Id)
		require.Equal(t, bamiyan.Name, friends.Another.Name)
		require.Equal(t, bamiyan.Name, friends.Another.Name)

		// update
		result, err = sess.Update("dbr_people").Where(Eq("id", jonathan.Id)).Set("name", "jonathan1").Exec()
		require.NoError(t, err)

		rowsAffected, err = result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		var n NullInt64
		sess.Select("count(*)").From("dbr_people").Where("name = ?", "jonathan1").LoadOne(&n)
		require.Equal(t, int64(1), n.Int64)

		// delete
		result, err = sess.DeleteFrom("dbr_people").Where(Eq("id", jonathan.Id)).Exec()
		require.NoError(t, err)

		rowsAffected, err = result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), rowsAffected)

		// select id
		ids, err = sess.Select("id").From("dbr_people").ReturnInt64s()
		require.NoError(t, err)
		require.Equal(t, 1, len(ids))
	}
}

func TestTimeout(t *testing.T) {
	mysqlSession := createSession("mysql", mysqlDSN)
	postgresSession := createSession("postgres", postgresDSN)
	sqlite3Session := createSession("sqlite3", sqlite3DSN)

	// all test sessions should be here
	testSession := []*Session{mysqlSession, postgresSession, sqlite3Session}

	for _, sess := range testSession {
		reset(t, sess)

		// session op timeout
		sess.Timeout = time.Nanosecond
		var people []dbrPerson
		_, err := sess.Select("*").From("dbr_people").Load(&people)
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, 1, sess.EventReceiver.(*testTraceReceiver).errored)

		_, err = sess.InsertInto("dbr_people").Columns("name", "email").Values("test", "test@test.com").Exec()
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, 2, sess.EventReceiver.(*testTraceReceiver).errored)

		_, err = sess.Update("dbr_people").Set("name", "test1").Exec()
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, 3, sess.EventReceiver.(*testTraceReceiver).errored)

		_, err = sess.DeleteFrom("dbr_people").Exec()
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, 4, sess.EventReceiver.(*testTraceReceiver).errored)

		// tx op timeout
		sess.Timeout = 0
		tx, err := sess.Begin()
		require.NoError(t, err)
		defer tx.RollbackUnlessCommitted()
		tx.Timeout = time.Nanosecond

		_, err = tx.Select("*").From("dbr_people").Load(&people)
		require.Equal(t, context.DeadlineExceeded, err)

		_, err = tx.InsertInto("dbr_people").Columns("name", "email").Values("test", "test@test.com").Exec()
		require.Equal(t, context.DeadlineExceeded, err)

		_, err = tx.Update("dbr_people").Set("name", "test1").Exec()
		require.Equal(t, context.DeadlineExceeded, err)

		_, err = tx.DeleteFrom("dbr_people").Exec()
		require.Equal(t, context.DeadlineExceeded, err)
	}
}
