package dbr

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

//
// Test helpers
//

// Returns a session that's not backed by a database
func createFakeSession() *Session {
	cxn := NewConnection(nil, nil)
	return cxn.NewSession(nil)
}

func createRealSession() *Session {
	cxn := NewConnection(realDb(), nil)
	return cxn.NewSession(nil)
}

func createRealSessionWithFixtures() *Session {
	sess := createRealSession()
	installFixtures(sess.cxn.Db)
	return sess
}

func realDb() *sql.DB {
	driver := os.Getenv("DBR_TEST_DRIVER")
	if driver == "" {
		driver = "mysql"
	}

	dsn := os.Getenv("DBR_TEST_DSN")
	if dsn == "" {
		dsn = "root:unprotected@unix(/tmp/mysql.sock)/uservoice_development?charset=utf8&parseTime=true"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalln("Mysql error ", err)
	}

	return db
}

type dbrPerson struct {
	Id    int64
	Name  string
	Email NullString
	Key   NullString
}

type nullTypedRecord struct {
	Id         int64
	StringVal  NullString
	Int64Val   NullInt64
	Float64Val NullFloat64
	TimeVal    NullTime
	BoolVal    NullBool
}

func installFixtures(db *sql.DB) {
	createPeopleTable := fmt.Sprintf(`
		CREATE TABLE dbr_people (
			id int(11) DEFAULT NULL auto_increment PRIMARY KEY,
			name varchar(255) NOT NULL,
			email varchar(255),
			%s varchar(255)
		)
	`, "`key`")

	createNullTypesTable := `
		CREATE TABLE null_types (
			id int(11) DEFAULT NULL auto_increment PRIMARY KEY,
			string_val varchar(255) NULL,
			int64_val int(11) NULL,
			float64_val float NULL,
			time_val datetime NULL,
			bool_val bool NULL
		)
	`

	sqlToRun := []string{
		"DROP TABLE IF EXISTS dbr_people",
		createPeopleTable,
		"INSERT INTO dbr_people (name,email) VALUES ('Jonathan', 'jonathan@uservoice.com')",
		"INSERT INTO dbr_people (name,email) VALUES ('Dmitri', 'zavorotni@jadius.com')",

		"DROP TABLE IF EXISTS null_types",
		createNullTypesTable,
	}

	for _, v := range sqlToRun {
		_, err := db.Exec(v)
		if err != nil {
			log.Fatalln("Failed to execute statement: ", v, " Got error: ", err)
		}
	}
}
