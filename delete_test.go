package dbr

import (
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestDeleteStmt(t *testing.T) {
	buf := NewBuffer()
	builder := DeleteFrom("table").Where(Eq("a", 1)).Comment("DELETE TEST")
	err := builder.Build(dialect.MySQL, buf)
	require.NoError(t, err)
	require.Equal(t, "/* DELETE TEST */\nDELETE FROM `table` WHERE (`a` = ?)", buf.String())
	require.Equal(t, []interface{}{1}, buf.Value())
}

func BenchmarkDeleteSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		DeleteFrom("table").Where(Eq("a", 1)).Build(dialect.MySQL, buf)
	}
}
