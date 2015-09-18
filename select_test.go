package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestSelectStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Select("a", "b").
		From(Select("a").From("table")).
		LeftJoin("table2", "table.a1 = table.a2").
		Distinct().
		Where(Eq("c", 1)).
		GroupBy("d").
		Having(Eq("e", 2)).
		OrderAsc("f").
		Limit(3).
		Offset(4)
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT a, b FROM ? LEFT JOIN `table2` ON table.a1 = table.a2 WHERE (`c` = ?) GROUP BY d HAVING (`e` = ?) ORDER BY f ASC LIMIT 3 OFFSET 4", buf.String())
	// two functions cannot be compared
	assert.Equal(t, 3, len(buf.Value()))
}

func BenchmarkSelectSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Select("a", "b").From("table").Where(Eq("c", 1)).OrderAsc("d").Build(dialect.MySQL, buf)
	}
}
