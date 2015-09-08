package dbr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkDeleteSql(b *testing.B) {
	s := createFakeSession()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.DeleteFrom("alpha").Where("a", "b").Limit(1).OrderDir("id", true).ToSql()
	}
}

func TestDeleteAllToSql(t *testing.T) {
	s := createFakeSession()

	sql, _ := s.DeleteFrom("a").ToSql()

	assert.Equal(t, sql, "DELETE FROM a")
}

func TestDeleteSingleToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.DeleteFrom("a").Where("id = ?", 1).ToSql()

	assert.Equal(t, sql, "DELETE FROM a WHERE (id = ?)")
	assert.Equal(t, args, []interface{}{1})
}

func TestDeleteTenStaringFromTwentyToSql(t *testing.T) {
	s := createFakeSession()

	sql, _ := s.DeleteFrom("a").Limit(10).Offset(20).OrderBy("id").ToSql()

	assert.Equal(t, sql, "DELETE FROM a ORDER BY id LIMIT 10 OFFSET 20")
}

func TestDeleteReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Insert a Barack
	res, err := s.InsertInto("dbr_people").Columns("name", "email").Values("Barack", "barack@whitehouse.gov").Exec()
	assert.NoError(t, err)

	// Get Barack's ID
	id, err := res.LastInsertId()
	assert.NoError(t, err)

	// Delete Barack
	res, err = s.DeleteFrom("dbr_people").Where("id = ?", id).Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, int64(1))

	var count int64
	err = s.Select("count(*)").From("dbr_people").Where("id = ?", id).LoadValue(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, int64(0))
}
