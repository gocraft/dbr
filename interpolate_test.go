package dbr

import (
	"testing"
	"time"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestInterpolateForDialect(t *testing.T) {
	for _, test := range []struct {
		query string
		value []interface{}
		want  string
	}{
		{
			query: "?",
			value: []interface{}{nil},
			want:  "NULL",
		},
		{
			query: "?",
			value: []interface{}{`'"'"`},
			want:  "'\\'\\\"\\'\\\"'",
		},
		{
			query: "? ?",
			value: []interface{}{true, false},
			want:  "1 0",
		},
		{
			query: "? ?",
			value: []interface{}{1, 1.23},
			want:  "1 1.23",
		},
		{
			query: "?",
			value: []interface{}{time.Date(2008, 9, 17, 20, 4, 26, 0, time.UTC)},
			want:  "'2008-09-17 20:04:26'",
		},
		{
			query: "?",
			value: []interface{}{[]string{"one", "two"}},
			want:  "('one','two')",
		},
		{
			query: "?",
			value: []interface{}{[]byte{0x1, 0x2, 0x3}},
			want:  "0x010203",
		},
		{
			query: "start?end",
			value: []interface{}{new(int)},
			want:  "start0end",
		},
		{
			query: "?",
			value: []interface{}{Select("a").From("table")},
			want:  "(SELECT a FROM table)",
		},
		{
			query: "?",
			value: []interface{}{I("a1").As("a2")},
			want:  "`a1` AS `a2`",
		},
		{
			query: "?",
			value: []interface{}{Select("a").From("table").As("a1")},
			want:  "(SELECT a FROM table) AS `a1`",
		},
		{
			query: "?",
			value: []interface{}{
				UnionAll(
					Select("a").From("table1"),
					Select("b").From("table2"),
				).As("t"),
			},
			want: "((SELECT a FROM table1) UNION ALL (SELECT b FROM table2)) AS `t`",
		},
	} {
		s, err := InterpolateForDialect(test.query, test.value, dialect.MySQL)
		assert.NoError(t, err)
		assert.Equal(t, test.want, s)
	}
}
