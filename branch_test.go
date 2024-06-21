package dbr

import (
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestWhen(t *testing.T) {
	for _, test := range []struct {
		when  Builder
		query string
		value []interface{}
	}{
		{
			when:  When(Eq("col", 1), 1),
			query: "WHEN (`col` = ?) THEN ?",
			value: []interface{}{1, 1},
		},
		{
			when:  When(And(Gt("a", 1), Lt("b", 2)), "c"),
			query: "WHEN ((`a` > ?) AND (`b` < ?)) THEN ?",
			value: []interface{}{1, 2, "c"},
		},
		{
			when:  When(Eq("a", 1), Gt("b", 2)),
			query: "WHEN (`a` = ?) THEN `b` > ?",
			value: []interface{}{1, 2},
		},
	} {
		buf := NewBuffer()
		err := test.when.Build(dialect.ClickHouse, buf)
		require.NoError(t, err)
		require.Equal(t, test.query, buf.String())
		require.Equal(t, test.value, buf.Value())
	}
}

func TestCase(t *testing.T) {
	for _, test := range []struct {
		caseBuilder Builder
		query       string
		value       []interface{}
	}{
		{
			caseBuilder: Case(When(Eq("col", 1), 1)),
			query:       "CASE WHEN (`col` = ?) THEN ? END",
			value:       []interface{}{1, 1},
		},
		{
			caseBuilder: Case(When(Eq("col", 1), 1)).As("a"),
			query:       "CASE WHEN (`col` = ?) THEN ? END AS `a`",
			value:       []interface{}{1, 1},
		},
		{
			caseBuilder: Case(When(Eq("col", 1), 2), Else(3)),
			query:       "CASE WHEN (`col` = ?) THEN ? ELSE ? END",
			value:       []interface{}{1, 2, 3},
		},
		{
			caseBuilder: Case(When(Eq("a", 1), 2), Else(Gt("b", 3))),
			query:       "CASE WHEN (`a` = ?) THEN ? ELSE `b` > ? END",
			value:       []interface{}{1, 2, 3},
		},
		{
			caseBuilder: Case(When(Eq("col", "a"), 1), When(Eq("col", "b"), 2), Else(3)),
			query:       "CASE WHEN (`col` = ?) THEN ? WHEN (`col` = ?) THEN ? ELSE ? END",
			value:       []interface{}{"a", 1, "b", 2, 3},
		},
		{
			caseBuilder: Case(When(Eq("colA", "a"), Lt("colB", 5)), When(Eq("colA", "b"), Lt("colB", 10)), Else(Lt("colB", 15))),
			query:       "CASE WHEN (`colA` = ?) THEN `colB` < ? WHEN (`colA` = ?) THEN `colB` < ? ELSE `colB` < ? END",
			value:       []interface{}{"a", 5, "b", 10, 15},
		},
		{
			caseBuilder: Case(When(Eq("colA", "a"), Lt("colB", 5)), When(Eq("colA", "b"), Lt("colB", 10)), Else(Lt("colB", 15))).As("c"),
			query:       "CASE WHEN (`colA` = ?) THEN `colB` < ? WHEN (`colA` = ?) THEN `colB` < ? ELSE `colB` < ? END AS `c`",
			value:       []interface{}{"a", 5, "b", 10, 15},
		},
	} {
		buf := NewBuffer()
		err := test.caseBuilder.Build(dialect.ClickHouse, buf)
		require.NoError(t, err)
		require.Equal(t, test.query, buf.String())
		require.Equal(t, test.value, buf.Value())
	}
}
