package dbr

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/gocraft/dbr/v2/dialect"
)

func TestSelectStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Select("a", "b").
		From(Select("a").From("table")).
		LeftJoin("table2", "table.a1 = table.a2", UseIndex("idx_table2")).
		Distinct().
		Where(Eq("c", 1)).
		GroupBy("d").
		Having(Eq("e", 2)).
		OrderAsc("f").
		Limit(3).
		Offset(4).
		Suffix("FOR UPDATE").
		Comment("SELECT TEST").
		IndexHint(UseIndex("idx_c_d").ForGroupBy(), "USE INDEX(idx_e_f)").
		IndexHint(IgnoreIndex("idx_a_b")).
		OptimizerHint("NO_RANGE_OPTIMIZATION(table)").
		OptimizerHint("MAX_EXECUTION_TIME(1000)")

	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "/* SELECT TEST */\nSELECT /*+ NO_RANGE_OPTIMIZATION(table) MAX_EXECUTION_TIME(1000) */ DISTINCT a, b FROM ? USE INDEX FOR GROUP BY(`idx_c_d`) USE INDEX(idx_e_f) IGNORE INDEX(`idx_a_b`) "+
		"LEFT JOIN `table2` USE INDEX(`idx_table2`) ON table.a1 = table.a2 WHERE (`c` = ?) GROUP BY d HAVING (`e` = ?) ORDER BY f ASC LIMIT 3 OFFSET 4 FOR UPDATE", buf.String())
	// two functions cannot be compared
	require.Equal(t, 3, len(buf.Value()))
}

func BenchmarkSelectSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Select("a", "b").From("table").Where(Eq("c", 1)).OrderAsc("d").Build(dialect.MySQL, buf)
	}
}

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
		require.NoError(t, err)

		//plain string slice (original behavior)
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
		require.NoError(t, err)

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

func TestSelectRows(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		_, err := sess.InsertInto("dbr_people").
			Columns("name", "email").
			Values("test1", "test1@test.com").
			Values("test2", "test2@test.com").
			Values("test3", "test3@test.com").
			Exec()
		require.NoError(t, err)

		rows, err := sess.Select("*").From("dbr_people").OrderAsc("id").Rows()
		require.NoError(t, err)
		defer rows.Close()

		want := []dbrPerson{
			{Id: 1, Name: "test1", Email: "test1@test.com"},
			{Id: 2, Name: "test2", Email: "test2@test.com"},
			{Id: 3, Name: "test3", Email: "test3@test.com"},
		}

		count := 0
		for rows.Next() {
			var p dbrPerson
			require.NoError(t, rows.Scan(&p.Id, &p.Name, &p.Email))
			require.Equal(t, want[count], p)
			count++
		}

		require.Equal(t, len(want), count)
	}
}

func TestInterfaceLoader(t *testing.T) {
	for _, sess := range testSession {
		reset(t, sess)

		_, err := sess.InsertInto("dbr_people").
			Columns("name", "email").
			Values("test1", "test1@test.com").
			Values("test2", "test2@test.com").
			Values("test2", "test3@test.com").
			Exec()
		require.NoError(t, err)

		var m []interface{}
		cnt, err := sess.Select("*").From("dbr_people").Load(InterfaceLoader(&m, dbrPerson{}))
		require.NoError(t, err)
		require.Equal(t, 3, cnt)
		require.Len(t, m, 3)
		person, ok := m[0].(dbrPerson)
		require.True(t, ok)
		require.Equal(t, "test1", person.Name)
	}
}

func TestPostgresArray(t *testing.T) {
	sess := postgresSession
	for _, v := range []string{
		`DROP TABLE IF EXISTS array_table`,
		`CREATE TABLE array_table (
			val integer[]
		)`,
	} {
		_, err := sess.Exec(v)
		require.NoError(t, err)
	}

	// INSERT INTO "array_table" ("val") VALUES ('{1,2,3}')
	_, err := sess.InsertInto("array_table").
		Pair("val", pq.Array([]int64{1, 2, 3})).
		Exec()
	require.NoError(t, err)

	var ns []int64
	err = sess.Select("val").From("array_table").LoadOne(pq.Array(&ns))
	require.NoError(t, err)

	require.Equal(t, []int64{1, 2, 3}, ns)
}
