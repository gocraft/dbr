package dbr

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		s.Insert("alpha").Columns("something_id", "user_id", "other").Values(1, 2, true).ToSql()
	}
}

func BenchmarkInsertRecordsSql(b *testing.B) {
	s := createFakeSession()
	obj := someRecord{1, 99, false}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Insert("alpha").Columns("something_id", "user_id", "other").Record(obj).ToSql()
	}
}

func TestInsertSingleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Insert("a").Columns("b", "c").Values(1, 2).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (b,c) VALUES (?,?)")
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestInsertMultipleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Insert("a").Columns("b", "c").Values(1, 2).Values(3, 4).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (b,c) VALUES (?,?),(?,?)")
	assert.Equal(t, args, []interface{}{1, 2, 3, 4})
}

func TestInsertRecordsToSql(t *testing.T) {
	s := createFakeSession()

	objs := []someRecord{{1, 88, false}, {2, 99, true}}
	sql, args := s.Insert("a").Columns("something_id", "user_id", "other").Record(objs[0]).Record(objs[1]).ToSql()

	assert.Equal(t, sql, "INSERT INTO a (something_id,user_id,other) VALUES (?,?,?),(?,?,?)")
	assert.Equal(t, args, []interface{}{1, 88, false, 2, 99, true})
}

func TestInsertReal(t *testing.T) {
	s := createRealSessionWithFixtures()
	_, e := s.Insert("dbr_people").Columns("name", "email").Values("Barack", "obama@whitehouse.com").Exec()

	assert.NoError(t, e)
	// id, err := r.LastInsertId()
	// 	rowsAff, err := r.RowsAffected()
	// 	println(id, err, rowsAff)

	// TODO: do a Query to get the result back
}

// TODO: test that we can use a record and it sets the ID
