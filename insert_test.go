package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
	"fmt"
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
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `table` (`a`,`b`) VALUES (?,?), (?,?)", buf.String())
	assert.Equal(t, []interface{}{1, "one", 2, "two"}, buf.Value())
}

func TestUpsertStmt(t *testing.T) {

	setMap := map[string]interface{}{"b": "three"}
	fmt.Println(setMap)

	buf := NewBuffer()
	builder := InsertInto("table").Columns("a", "b").Values(1, "one").Record(&insertTest{
		A: 2,
		C: "two",
	}).OnDuplicateKeyUpdate(setMap)
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `table` (`a`,`b`) VALUES (?,?), (?,?) ON DUPLICATE KEY UPDATE `b` = ?", buf.String())
	assert.Equal(t, []interface{}{1, "one", 2, "two", "three"}, buf.Value())
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
