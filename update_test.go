package dbr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkUpdateValuesSql(b *testing.B) {
	s := createFakeSession()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Update("alpha").Set("something_id", 1).Where("id", 1).ToSql()
	}
}

func BenchmarkUpdateValueMapSql(b *testing.B) {
	s := createFakeSession()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Update("alpha").Set("something_id", 1).SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id", 1).ToSql()
	}
}

func TestUpdateAllToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Update("a").Set("b", 1).Set("c", 2).ToSql()

	assert.Equal(t, sql, "UPDATE a SET `b` = ?, `c` = ?")
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestUpdateSingleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Update("a").Set("b", 1).Set("c", 2).Where("id = ?", 1).ToSql()

	assert.Equal(t, sql, "UPDATE a SET `b` = ?, `c` = ? WHERE (id = ?)")
	assert.Equal(t, args, []interface{}{1, 2, 1})
}

func TestUpdateSetMapToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Update("a").SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id = ?", 1).ToSql()

	if sql == "UPDATE a SET `b` = ?, `c` = ? WHERE (id = ?)" {
		assert.Equal(t, args, []interface{}{1, 2, 1})
	} else {
		assert.Equal(t, sql, "UPDATE a SET `c` = ?, `b` = ? WHERE (id = ?)")
		assert.Equal(t, args, []interface{}{2, 1, 1})
	}
}

func TestUpdateSetExprToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + 1")).Where("id = ?", 9).ToSql()

	assert.Equal(t, sql, "UPDATE a SET `foo` = ?, `bar` = COALESCE(bar, 0) + 1 WHERE (id = ?)")
	assert.Equal(t, args, []interface{}{1, 9})

	sql, args = s.Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + ?", 2)).Where("id = ?", 9).ToSql()

	assert.Equal(t, sql, "UPDATE a SET `foo` = ?, `bar` = COALESCE(bar, 0) + ? WHERE (id = ?)")
	assert.Equal(t, args, []interface{}{1, 2, 9})
}

func TestUpdateTenStaringFromTwentyToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Update("a").Set("b", 1).Limit(10).Offset(20).ToSql()

	assert.Equal(t, sql, "UPDATE a SET `b` = ? LIMIT 10 OFFSET 20")
	assert.Equal(t, args, []interface{}{1})
}

func TestUpdateKeywordColumnName(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Insert a user with a key
	res, err := s.InsertInto("dbr_people").Columns("name", "email", "key").Values("Benjamin", "ben@whitehouse.gov", "6").Exec()
	assert.NoError(t, err)

	// Update the key
	res, err = s.Update("dbr_people").Set("key", "6-revoked").Where(Eq{"key": "6"}).Exec()
	assert.NoError(t, err)

	// Assert our record was updated (and only our record)
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)

	var person dbrPerson
	err = s.Select("*").From("dbr_people").Where(Eq{"email": "ben@whitehouse.gov"}).LoadOne(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Name, "Benjamin")
	assert.Equal(t, person.Key.String, "6-revoked")
}

func TestUpdateReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Insert a George
	res, err := s.InsertInto("dbr_people").Columns("name", "email").Values("George", "george@whitehouse.gov").Exec()
	assert.NoError(t, err)

	// Get George's ID
	id, err := res.LastInsertId()
	assert.NoError(t, err)

	// Rename our George to Barack
	res, err = s.Update("dbr_people").SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).Where("id = ?", id).Exec()

	assert.NoError(t, err)

	var person dbrPerson
	err = s.Select("*").From("dbr_people").Where("id = ?", id).LoadOne(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Id, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}
