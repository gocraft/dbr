package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestCondition(t *testing.T) {
	for _, test := range []struct {
		cond  Condition
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
	} {
		buf := NewBuffer()
		err := test.cond.Build(dialect.MySQL, buf)
		assert.NoError(t, err)
		assert.Equal(t, test.query, buf.String())
		assert.Equal(t, test.value, buf.Value())
	}
}
