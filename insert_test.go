package dbr

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

type someRecord struct {
	SomethingId int
	UserId      int64
	Other       bool
}

func BenchmarkInsertValuesSql(b *testing.B) {
	s := createFakeSession()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.InsertInto("alpha").Columns("something_id", "user_id", "other").Values(1, 2, true).ToSql()
	}
}

func BenchmarkInsertRecordsSql(b *testing.B) {
	s := createFakeSession()
	obj := someRecord{1, 99, false}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.InsertInto("alpha").Columns("something_id", "user_id", "other").Record(obj).ToSql()
	}
}

func TestInsertSingleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.InsertInto("a").Columns("b", "c").Values(1, 2).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (`b`,`c`) VALUES (?,?)")
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestInsertMultipleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.InsertInto("a").Columns("b", "c").Values(1, 2).Values(3, 4).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (`b`,`c`) VALUES (?,?),(?,?)")
	assert.Equal(t, args, []interface{}{1, 2, 3, 4})
}

func TestInsertRecordsToSql(t *testing.T) {
	s := createFakeSession()

	objs := []someRecord{{1, 88, false}, {2, 99, true}}
	sql, args := s.InsertInto("a").Columns("something_id", "user_id", "other").Record(objs[0]).Record(objs[1]).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (`something_id`,`user_id`,`other`) VALUES (?,?,?),(?,?,?)")
	assert.Equal(t, args, []interface{}{1, 88, false, 2, 99, true})
}

func TestInsertKeywordColumnName(t *testing.T) {
	// Insert a column whose name is reserved
	s := createRealSessionWithFixtures()
	res, err := s.InsertInto("dbr_people").Columns("name", "key").Values("Barack", "44").Exec()
	assert.NoError(t, err)

	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)
}

func TestInsertReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()
	res, err := s.InsertInto("dbr_people").Columns("name", "email").Values("Barack", "obama@whitehouse.gov").Exec()
	validateInsertingBarack(t, s, res, err)

	// Insert by specifying a record (ptr to struct)
	s = createRealSessionWithFixtures()
	person := dbrPerson{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama@whitehouse.gov"
	res, err = s.InsertInto("dbr_people").Columns("name", "email").Record(&person).Exec()
	validateInsertingBarack(t, s, res, err)

	// Insert by specifying a record (struct)
	s = createRealSessionWithFixtures()
	res, err = s.InsertInto("dbr_people").Columns("name", "email").Record(person).Exec()
	validateInsertingBarack(t, s, res, err)
}

func validateInsertingBarack(t *testing.T, s *Session, res sql.Result, err error) {
	assert.NoError(t, err)
	id, err := res.LastInsertId()
	assert.NoError(t, err)
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)

	assert.True(t, id > 0)
	assert.Equal(t, rowsAff, 1)

	var person dbrPerson
	err = s.Select("*").From("dbr_people").Where("id = ?", id).LoadStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Id, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "obama@whitehouse.gov")
}

// TODO: do a real test inserting multiple records
