package dbr

import (
	"testing"

	"github.com/stretchr/testify/require"
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

		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, stringSlice, 3)

		//string slice with sql.Scanner implemented, should act as a single record
		var sliceScanner stringSliceWithSQLScanner
		cnt, err = sess.Select("name").From("dbr_people").Load(&sliceScanner)

		require.NoError(t, err)
		require.Equal(t, 1, cnt)
		require.Len(t, sliceScanner, 1)
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
		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, m, 3)
		require.Equal(t, "test1", m["test1@test.com"])

		var m2 map[int64]*dbrPerson
		cnt, err = sess.Select("id, name, email").From("dbr_people").Load(&m2)
		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, m2, 3)
		require.Equal(t, "test1@test.com", m2[1].Email)
		require.Equal(t, "test1", m2[1].Name)
		// the id value is used as the map key, so it is not hydrated in the struct
		require.Equal(t, int64(0), m2[1].Id)

		var m3 map[string][]string
		cnt, err = sess.Select("name, email").From("dbr_people").OrderAsc("id").Load(&m3)
		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, m3, 2)
		require.Equal(t, []string{"test1@test.com"}, m3["test1"])
		require.Equal(t, []string{"test2@test.com", "test3@test.com"}, m3["test2"])

		var set map[string]struct{}
		cnt, err = sess.Select("name").From("dbr_people").Load(&set)
		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, set, 2)
		_, ok := set["test1"]
		require.True(t, ok)
	}
}
