package dbr

import (
	"testing"

	"dbr-aaa/vendor/github.com/gocraft/dbr/dialect"
	"dbr-aaa/vendor/github.com/stretchr/testify/require"
)

func TestUpdateStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Update("table").Set("a", 1).Where(Eq("b", 2))
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)

	require.Equal(t, "UPDATE `table` SET `a` = ? WHERE (`b` = ?)", buf.String())
	require.Equal(t, []interface{}{1, 2}, buf.Value())
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
