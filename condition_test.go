package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestCondition(t *testing.T) {
	for _, test := range []struct {
		cond  Builder
		query string
		value []interface{}
		isErr bool
	}{
		{
			cond:  Eq("col", 1),
			query: "`col` = ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Eq("col", nil),
			query: "`col` IS NULL",
			value: nil,
			isErr: false,
		},
		{
			cond:  Eq("col", []int{}),
			query: "0",
			value: nil,
			isErr: false,
		},
		{
			cond:  Neq("col", 1),
			query: "`col` != ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Neq("col", nil),
			query: "`col` IS NOT NULL",
			value: nil,
			isErr: false,
		},
		{
			cond:  Gt("col", 1),
			query: "`col` > ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Gte("col", 1),
			query: "`col` >= ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Lt("col", 1),
			query: "`col` < ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Lte("col", 1),
			query: "`col` <= ?",
			value: []interface{}{1},
			isErr: false,
		},
		{
			cond:  Like("col", 1),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  Like("col", "like"),
			query: "`col` LIKE ?",
			value: []interface{}{"like"},
			isErr: false,
		},
		{
			cond:  Like("col", []rune{'l', 'i', 'k', 'e'}),
			query: "`col` LIKE ?",
			value: []interface{}{"like"},
			isErr: false,
		},
		{
			cond:  Like("col", []byte("like")),
			query: "`col` LIKE ?",
			value: []interface{}{"like"},
			isErr: false,
		},
		{
			cond:  Like("col", []int{}),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  Like("col", nil),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  NotLike("col", 1),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  NotLike("col", "not like"),
			query: "`col` NOT LIKE ?",
			value: []interface{}{"not like"},
			isErr: false,
		},
		{
			cond:  NotLike("col", []rune{'n', 'o', 't', ' ', 'l', 'i', 'k', 'e'}),
			query: "`col` NOT LIKE ?",
			value: []interface{}{"not like"},
			isErr: false,
		},
		{
			cond:  NotLike("col", []byte("not like")),
			query: "`col` NOT LIKE ?",
			value: []interface{}{"not like"},
			isErr: false,
		},
		{
			cond:  NotLike("col", []int{}),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  NotLike("col", nil),
			query: "",
			value: nil,
			isErr: true,
		},
		{
			cond:  And(Lt("a", 1), Or(Gt("b", 2), Neq("c", 3))),
			query: "(`a` < ?) AND ((`b` > ?) OR (`c` != ?))",
			value: []interface{}{1, 2, 3},
			isErr: false,
		},
	} {
		buf := NewBuffer()
		err := test.cond.Build(dialect.MySQL, buf)
		if !test.isErr {
			assert.NoError(t, err)
		}
		assert.Equal(t, test.query, buf.String())
		assert.Equal(t, test.value, buf.Value())
	}
}
