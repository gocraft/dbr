package dbr

import (
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const (
	mysqlDSN    = "root@unix(/tmp/mysql.sock)/uservoice_test?charset=utf8"
	postgresDSN = "postgres://postgres@localhost:5432/uservoice_test?sslmode=disable"
)

func createSession(driver, dsn string) *Session {
	var testDSN string
	switch driver {
	case "mysql":
		testDSN = os.Getenv("DBR_TEST_MYSQL_DSN")
	case "postgres":
		testDSN = os.Getenv("DBR_TEST_POSTGRES_DSN")
	}
	if testDSN != "" {
		dsn = testDSN
	}
	conn, err := Open(driver, dsn, nil)
	if err != nil {
		log.Fatal(err)
	}
	reset(conn)
	sess := conn.NewSession(nil)
	return sess
}

var (
	mysqlSession    = createSession("mysql", mysqlDSN)
	postgresSession = createSession("postgres", postgresDSN)
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

func reset(conn *Connection) {
	// serial = BIGINT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE
	// the following sql should work for both mysql and postgres
	for _, v := range []string{
		`DROP TABLE IF EXISTS dbr_people`,
		`CREATE TABLE dbr_people (
			id serial PRIMARY KEY,
			name varchar(255) NOT NULL,
			email varchar(255)
		)`,

		`DROP TABLE IF EXISTS null_types`,
		`CREATE TABLE null_types (
			id serial PRIMARY KEY,
			string_val varchar(255) NULL,
			int64_val integer NULL,
			float64_val float NULL,
			time_val timestamp NULL ,
			bool_val bool NULL
		)`,
	} {
		_, err := conn.Exec(v)
		if err != nil {
			log.Fatalf("Failed to execute statement: %s, Got error: %s", v, err)
		}
	}
}

func TestBasicCRUD(t *testing.T) {
	jonathan := dbrPerson{
		Name:  "jonathan",
		Email: "jonathan@uservoice.com",
	}
	dmitri := dbrPerson{
		Name:  "dmitri",
		Email: "zavorotni@jadius.com",
	}
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		// insert
		result, err := sess.InsertInto("dbr_people").Columns("name", "email").Record(&jonathan).Exec()
		assert.NoError(t, err)
		assert.True(t, jonathan.Id > 0)

		result, err = sess.InsertInto("dbr_people").Columns("name", "email").Record(&dmitri).Exec()
		assert.NoError(t, err)
		assert.True(t, dmitri.Id > 0)

		// select
		var people []dbrPerson
		count, err := sess.Select("*").From("dbr_people").Where(Eq("id", jonathan.Id)).LoadStructs(&people)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, jonathan.Id, people[0].Id)
		assert.Equal(t, jonathan.Name, people[0].Name)
		assert.Equal(t, jonathan.Email, people[0].Email)

		// select id limit
		ids, err := sess.Select("id").From("dbr_people").Limit(1).ReturnInt64s()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ids))

		// update
		result, err = sess.Update("dbr_people").Where(Eq("id", jonathan.Id)).Set("name", "jonathan1").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		var n NullInt64
		sess.Select("count(*)").From("dbr_people").Where("name = ?", "jonathan1").LoadValue(&n)
		assert.EqualValues(t, 1, n.Int64)

		// delete
		result, err = sess.DeleteFrom("dbr_people").Where(Eq("id", jonathan.Id)).Exec()
		assert.NoError(t, err)

		rowsAffected, err = result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)
	}
}

func TestBatchInsert(t *testing.T) {
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		result, err := sess.InsertInto("dbr_people").Columns("name").Values("A").Values("B").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 2, rowsAffected)
	}
}

func TestInjectID(t *testing.T) {
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		var person dbrPerson
		result, err := sess.InsertInto("dbr_people").Columns("name").Record(&person).Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		assert.True(t, person.Id > 0)
	}
}

func TestNoInjectID(t *testing.T) {
	for _, sess := range []*Session{mysqlSession, postgresSession} {
		people := make([]dbrPerson, 2)
		result, err := sess.InsertInto("dbr_people").Columns("name").Record(&people[0]).Record(&people[1]).Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 2, rowsAffected)

		assert.EqualValues(t, 0, people[0].Id)
		assert.EqualValues(t, 0, people[1].Id)
	}
}
