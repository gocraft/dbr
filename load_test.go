package dbr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringSliceWithSQLScanner []string

func (ss *stringSliceWithSQLScanner) Scan(src interface{}) error {
	*ss = append(*ss, "called")
	return nil
}

func TestSliceWithSQLScannerSelect(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		_, err := sess.InsertInto("dbr_people").
			Columns("name", "email").
			Values("test1", "test1@test.com").
			Values("test2", "test2@test.com").
			Values("test3", "test3@test.com").
			Exec()

		//plain string slice (original behavour)
		var stringSlice []string
		cnt, err := sess.Select("name").From("dbr_people").Load(&stringSlice)

		assert.NoError(t, err)
		assert.Equal(t, cnt, 3)
		assert.Len(t, stringSlice, 3)

		//string slice with sql.Scanner implemented, should act as a single record
		var sliceScanner stringSliceWithSQLScanner
		cnt, err = sess.Select("name").From("dbr_people").Load(&sliceScanner)

		assert.NoError(t, err)
		assert.Equal(t, cnt, 1)
		assert.Len(t, sliceScanner, 1)
	}
}

func TestMaps(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		_, err := sess.InsertInto("dbr_people").
			Columns("name", "email").
			Values("test1", "test1@test.com").
			Values("test2", "test2@test.com").
			Values("test2", "test3@test.com").
			Exec()

		var m map[string]string
		cnt, err := sess.Select("email, name").From("dbr_people").Load(&m)
		assert.NoError(t, err)
		assert.Equal(t, cnt, 3)
		assert.Len(t, m, 3)
		assert.Equal(t, m["test1@test.com"], "test1")

		var m2 map[int64]*dbrPerson
		cnt, err = sess.Select("id, name, email").From("dbr_people").Load(&m2)
		assert.NoError(t, err)
		assert.Equal(t, cnt, 3)
		assert.Len(t, m2, 3)
		assert.Equal(t, m2[1].Email, "test1@test.com")
		assert.Equal(t, m2[1].Name, "test1")
		// the id value is used as the map key, so it is not hydrated in the struct
		assert.EqualValues(t, m2[1].Id, 0)

		var m3 map[string][]string
		cnt, err = sess.Select("name, email").From("dbr_people").OrderAsc("id").Load(&m3)
		assert.NoError(t, err)
		assert.Equal(t, cnt, 3)
		assert.Len(t, m3, 2)
		assert.Equal(t, m3["test1"], []string{"test1@test.com"})
		assert.Equal(t, m3["test2"], []string{"test2@test.com", "test3@test.com"})

		var set map[string]struct{}
		cnt, err = sess.Select("name").From("dbr_people").Load(&set)
		assert.NoError(t, err)
		assert.Equal(t, cnt, 3)
		assert.Len(t, set, 2)
		_, ok := set["test1"]
		assert.True(t, ok)
	}
}
