package dbr

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
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
	mysqlDSN      = os.Getenv("DBR_TEST_MYSQL_DSN")
	postgresDSN   = os.Getenv("DBR_TEST_POSTGRES_DSN")
	sqlite3DSN    = ":memory:"
	mssqlDSN      = os.Getenv("DBR_TEST_MSSQL_DSN")
	clickhouseDSN = os.Getenv("DBR_TEST_CLICKHOUSE_DSN")
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
	clickhouseSession     = createSession("clickhouse", clickhouseDSN)

	// all test sessions should be here
	testSession = []*Session{mysqlSession, postgresSession, sqlite3Session, mssqlSession, clickhouseSession}
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

func testSessionName(sess *Session) string {
	switch sess.Dialect {
	case dialect.MySQL:
		return "MySQL"
	case dialect.PostgreSQL:
		return "PostgreSQL"
	case dialect.SQLite3:
		return "SQLite3"
	case dialect.MSSQL:
		return "MSSQL"
	case dialect.Clickhouse:
		return "ClickHouse"
	}
	return ""
}

func dialectExpectedRowsAffected(sess *Session, expectedRows int64) int64 {
	if sess.Dialect == dialect.Clickhouse {
		// ClickHouse does not support returning rows affected, it always returns 0.
		return 0
	}
	return expectedRows
}

func reset(t *testing.T, sess *Session) {
	autoIncrementType := "serial PRIMARY KEY"
	boolType := "bool"
	datetimeType := "timestamp"
	engine := ""

	switch sess.Dialect {
	case dialect.SQLite3:
		autoIncrementType = "integer PRIMARY KEY"
	case dialect.MSSQL:
		autoIncrementType = "integer IDENTITY PRIMARY KEY"
		boolType = "BIT"
		datetimeType = "datetime"
	case dialect.Clickhouse:
		autoIncrementType = "integer"
		engine = "ENGINE = MergeTree ORDER BY id"
	}
	for _, v := range []string{
		`DROP TABLE IF EXISTS dbr_people`,
		fmt.Sprintf(`CREATE TABLE dbr_people (
			id %s,
			name varchar(255) NOT NULL,
			email varchar(255)
		) %s`, autoIncrementType, engine),

		`DROP TABLE IF EXISTS null_types`,
		fmt.Sprintf(`CREATE TABLE null_types (
			id %s,
			string_val varchar(255) NULL,
			int64_val integer NULL,
			float64_val float NULL,
			time_val %s NULL,
			bool_val %s NULL
		) %s`, autoIncrementType, datetimeType, boolType, engine),
	} {
		_, err := sess.Exec(v)
		require.NoError(t, err)
	}
	// clear test data collected by testTraceReceiver
	sess.EventReceiver = &testTraceReceiver{}
}

func TestBasicCRUD(t *testing.T) {
	for _, sess := range testSession {
		t.Run(testSessionName(sess), func(t *testing.T) {
			reset(t, sess)

			jonathan := dbrPerson{
				Name:  "jonathan",
				Email: "jonathan@uservoice.com",
			}
			insertColumns := []string{"name", "email"}
			if sess.Dialect == dialect.PostgreSQL || sess.Dialect == dialect.Clickhouse {
				jonathan.Id = 1
				insertColumns = []string{"id", "name", "email"}
			}
			if sess.Dialect == dialect.MSSQL {
				jonathan.Id = 1
			}

			// insert
			result, err := sess.InsertInto("dbr_people").Columns(insertColumns...).Record(&jonathan).Exec()
			require.NoError(t, err)

			rowsAffected, err := result.RowsAffected()
			require.NoError(t, err)
			require.Equal(t, dialectExpectedRowsAffected(sess, 1), rowsAffected)

			require.True(t, jonathan.Id > 0)
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
			require.Equal(t, 1, len(ids))

			// select id limit
			ids, err = sess.Select("id").From("dbr_people").Limit(1).ReturnInt64s()
			require.NoError(t, err)
			require.Equal(t, 1, len(ids))

			// update
			updateStmt := sess.Update("dbr_people").Where(Eq("id", jonathan.Id)).Set("name", "jonathan1")
			updateStmt.Settings("mutations_sync", "2") // This is for clickhouse to make sure the mutation is done before the select
			result, err = updateStmt.Exec()
			require.NoError(t, err)

			rowsAffected, err = result.RowsAffected()
			require.NoError(t, err)
			require.Equal(t, dialectExpectedRowsAffected(sess, 1), rowsAffected)

			var n NullInt64
			err = sess.Select("count(*)").From("dbr_people").Where("name = ?", "jonathan1").LoadOne(&n)
			require.NoError(t, err)
			require.Equal(t, int64(1), n.Int64)

			// delete
			result, err = sess.DeleteFrom("dbr_people").Where(Eq("id", jonathan.Id)).Exec()
			require.NoError(t, err)

			rowsAffected, err = result.RowsAffected()
			require.NoError(t, err)
			require.Equal(t, dialectExpectedRowsAffected(sess, 1), rowsAffected)

			// select id
			ids, err = sess.Select("id").From("dbr_people").ReturnInt64s()
			require.NoError(t, err)
			require.Equal(t, 0, len(ids))
		})
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

func TestOnConflict(t *testing.T) {
	for _, sess := range testSession {
		t.Run(testSessionName(sess), func(t *testing.T) {
			if sess.Dialect == dialect.SQLite3 || sess.Dialect == dialect.Clickhouse || sess.Dialect == dialect.MSSQL {
				t.Skip()
			}

			reset(t, sess)
			for i := 0; i < 2; i++ {
				b := sess.InsertInto("dbr_people").Columns("id", "name", "email").Values(1, "test", "test@test.com")
				b.OnConflict("dbr_people_pkey").Action("email", Expr("CONCAT(?, 2)", Proposed("email")))
				_, err := b.Exec()
				require.NoError(t, err)
			}
			var value string
			_, err := sess.SelectBySql("SELECT email FROM dbr_people WHERE id=?", "1").Load(&value)
			require.NoError(t, err)
			require.Equal(t, "test@test.com2", value)
		})
	}
}
