package dbr

import (
	"testing"

	"dbr/vendor/github.com/gocraft/dbr/dialect"
	"dbr/vendor/github.com/stretchr/testify/require"
)

func TestDeleteStmt(t *testing.T) {
	buf := NewBuffer()
	builder := DeleteFrom("table").Where(Eq("a", 1))
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "DELETE FROM `table` WHERE (`a` = ?)", buf.String())
	require.Equal(t, []interface{}{1}, buf.Value())
}

func BenchmarkDeleteSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		DeleteFrom("table").Where(Eq("a", 1)).Build(dialect.MySQL, buf)
	}
}
