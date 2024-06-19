package dbr

import (
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	for _, test := range []struct {
		cond  Builder
		query string
		value []interface{}
	}{
		{
			cond:  Eq("col", 1),
			query: "`col` = ?",
			value: []interface{}{1},
		},
		{
			cond:  Eq("col", nil),
			query: "`col` IS NULL",
			value: nil,
		},
		{
			cond:  Eq("col", []int{}),
			query: "0",
			value: nil,
		},
		{
			cond:  Neq("col", 1),
			query: "`col` != ?",
			value: []interface{}{1},
		},
		{
			cond:  Neq("col", nil),
			query: "`col` IS NOT NULL",
			value: nil,
		},
		{
			cond:  Gt("col", 1),
			query: "`col` > ?",
			value: []interface{}{1},
		},
		{
			cond:  Gte("col", 1),
			query: "`col` >= ?",
			value: []interface{}{1},
		},
		{
			cond:  Lt("col", 1),
			query: "`col` < ?",
			value: []interface{}{1},
		},
		{
			cond:  Lte("col", 1),
			query: "`col` <= ?",
			value: []interface{}{1},
		},
		{
			cond:  And(Lt("a", 1), Or(Gt("b", 2), Neq("c", 3))),
			query: "(`a` < ?) AND ((`b` > ?) OR (`c` != ?))",
			value: []interface{}{1, 2, 3},
		},
		{
			cond:  Like("a", "%BLAH%", "#"),
			query: "`a` LIKE '%BLAH%' ESCAPE '#'",
			value: nil,
		},
		{
			cond:  Like("a", "%50#%%", "#"),
			query: "`a` LIKE '%50#%%' ESCAPE '#'",
			value: nil,
		},
		{
			cond:  NotLike("a", "%BLAH%", "#"),
			query: "`a` NOT LIKE '%BLAH%' ESCAPE '#'",
			value: nil,
		},
		{
			cond:  NotLike("a", "%50#%%", "#"),
			query: "`a` NOT LIKE '%50#%%' ESCAPE '#'",
			value: nil,
		},
		{
			cond:  Like("a", "_x_"),
			query: "`a` LIKE '_x_'",
			value: nil,
		},
		{
			cond:  NotLike("a", "_x_"),
			query: "`a` NOT LIKE '_x_'",
			value: nil,
		},
	} {
		buf := NewBuffer()
		err := test.cond.Build(dialect.MySQL, buf)
		require.NoError(t, err)
		require.Equal(t, test.query, buf.String())
		require.Equal(t, test.value, buf.Value())
	}
}
