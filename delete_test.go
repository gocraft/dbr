package dbr

import (
	"testing"

	"github.com/mailru/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestDeleteStmt(t *testing.T) {
	buf := NewBuffer()
	builder := DeleteFrom("table").Where(Eq("a", 1))
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "DELETE FROM `table` WHERE (`a` = ?)", buf.String())
	assert.Equal(t, []interface{}{1}, buf.Value())
}

func BenchmarkDeleteSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		DeleteFrom("table").Where(Eq("a", 1)).Build(dialect.MySQL, buf)
	}
}
