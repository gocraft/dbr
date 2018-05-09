package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

type updateTest struct {
	C string `db:"b"`
	D string `db:"c"`
}

func TestUpdateStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Update("table").Set("a", 1).Columns([]string{"b"}...).Record(&updateTest{
		C: "two",
		D: "three",
	}).Where(Eq("b", 2))
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)

	assert.Equal(t, "UPDATE `table` SET `a` = ?, `b` = ? WHERE (`b` = ?)", buf.String())
	assert.Equal(t, []interface{}{1, "two", 2}, buf.Value())
}

func BenchmarkUpdateValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").Set("a", 1).Build(dialect.MySQL, buf)
	}
}

func BenchmarkUpdateMapSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").SetMap(map[string]interface{}{"a": 1, "b": 2}).Build(dialect.MySQL, buf)
	}
}

func BenchmarkUpdateRecordSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Update("table").Record(&updateTest{
			C: "two",
		}).Build(dialect.MySQL, buf)
	}
}
