package dbr

import (
	"testing"

	"github.com/mailru/dbr/dialect"
	"github.com/stretchr/testify/assert"
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

func TestInsertOnConflictStmt(t *testing.T) {
	buf := NewBuffer()
	exp := Expr("a + ?", 1)
	builder := InsertInto("table").Columns("a", "b").Values(1, "one")
	builder.OnConflict("").Action("a", exp).Action("b", "one")
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `table` (`a`,`b`) VALUES (?,?) ON DUPLICATE KEY UPDATE `a`=?,`b`=?", buf.String())
	assert.Equal(t, []interface{}{1, "one", exp, "one"}, buf.Value())
}

func TestInsertOnConflictMapStmt(t *testing.T) {
	buf := NewBuffer()
	exp := Expr("a + ?", 1)
	builder := InsertInto("table").Columns("a", "b").Values(1, "one")
	err := builder.OnConflictMap("", map[string]interface{}{"a": exp, "b": "one"}).Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `table` (`a`,`b`) VALUES (?,?) ON DUPLICATE KEY UPDATE `a`=?,`b`=?", buf.String())
	assert.Equal(t, []interface{}{1, "one", exp, "one"}, buf.Value())
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
