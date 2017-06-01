package dbr

import (
	"testing"

	"github.com/mailru/dbr/dialect"
	"github.com/stretchr/testify/assert"
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
			cond:  Eq("col", map[int]int{}),
			query: "0",
			value: nil,
		},
		{
			cond:  Eq("col", []int{1}),
			query: "`col` IN ?",
			value: []interface{}{[]int{1}},
		},
		{
			cond:  Eq("col", map[int]int{1: 2}),
			query: "`col` IN ?",
			value: []interface{}{map[int]int{1: 2}},
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
			cond:  Neq("col", []int{}),
			query: "1",
			value: nil,
		},
		{
			cond:  Neq("col", []int{1}),
			query: "`col` NOT IN ?",
			value: []interface{}{[]int{1}},
		},
		{
			cond:  Neq("col", map[int]int{1: 2}),
			query: "`col` NOT IN ?",
			value: []interface{}{map[int]int{1: 2}},
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
		buf := newBuffer()
		err := test.cond.Build(dialect.MySQL, buf)
		assert.NoError(t, err)
		assert.Equal(t, test.query, buf.String())
		assert.Equal(t, test.value, buf.Value())
	}
}
