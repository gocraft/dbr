package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/require"
)

type insertTest struct {
	A int
	C string `db:"b"`
}

func TestInsertStmt(t *testing.T) {
	buf := NewBuffer()
	builder := InsertInto("table").Columns("a", "b").Values(1, "one").Record(&insertTest{
		A: 2,
		C: "two",
	})
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "INSERT INTO `table` (`a`,`b`) VALUES (?,?), (?,?)", buf.String())
	require.Equal(t, []interface{}{1, "one", 2, "two"}, buf.Value())
}

func TestPostgresReturning(t *testing.T) {
	sess := postgresSession
	reset(t, sess)

	var person dbrPerson
	err := sess.InsertInto("dbr_people").Columns("name").Record(&person).
		Returning("id").Load(&person.Id)
	require.NoError(t, err)
	require.True(t, person.Id > 0)
}

func BenchmarkInsertValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b").Values(1, "one").Build(dialect.MySQL, buf)
	}
}

func BenchmarkInsertRecordSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b").Record(&insertTest{
			A: 2,
			C: "two",
		}).Build(dialect.MySQL, buf)
	}
}
