package dbr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestWhen(t *testing.T) {
	for _, test := range []struct {
		name  string
		when  Builder
		query string
		value []interface{}
	}{
		{
			name:  "when",
			when:  When(Eq("col", 1), 1),
			query: "WHEN (`col` = ?) THEN ?",
			value: []interface{}{1, 1},
		},
		{
			name:  "when with and",
			when:  When(And(Gt("a", 1), Lt("b", 2)), "c"),
			query: "WHEN ((`a` > ?) AND (`b` < ?)) THEN ?",
			value: []interface{}{1, 2, "c"},
		},
		{
			name:  "when eq then gt",
			when:  When(Eq("a", 1), Gt("b", 2)),
			query: "WHEN (`a` = ?) THEN `b` > ?",
			value: []interface{}{1, 2},
		},
	} {
		for _, sess := range testSession {
			t.Run(fmt.Sprintf("%s/%s", testSessionName(sess), test.name), func(t *testing.T) {
				reset(t, sess)

				expectedQuery := test.query
				if sess.Dialect != dialect.MySQL {
					// MySQL is the only one that uses a different QuoteIdent
					expectedQuery = strings.ReplaceAll(expectedQuery, "`", `"`)
				}

				buf := NewBuffer()
				err := test.when.Build(sess.Dialect, buf)
				require.NoError(t, err)
				require.Equal(t, expectedQuery, buf.String())
				require.Equal(t, test.value, buf.Value())
			})
		}
	}
}

func TestCase(t *testing.T) {
	for _, test := range []struct {
		name        string
		caseBuilder Builder
		query       string
		value       []interface{}
	}{
		{
			name:        "case",
			caseBuilder: Case(When(Eq("col", 1), 1)),
			query:       "CASE WHEN (`col` = ?) THEN ? END",
			value:       []interface{}{1, 1},
		},
		{
			name:        "case as",
			caseBuilder: Case(When(Eq("col", 1), 1)).As("a"),
			query:       "CASE WHEN (`col` = ?) THEN ? END AS `a`",
			value:       []interface{}{1, 1},
		},
		{
			name:        "case else",
			caseBuilder: Case(When(Eq("col", 1), 2), Else(3)),
			query:       "CASE WHEN (`col` = ?) THEN ? ELSE ? END",
			value:       []interface{}{1, 2, 3},
		},
		{
			name:        "case else with gt",
			caseBuilder: Case(When(Eq("a", 1), 2), Else(Gt("b", 3))),
			query:       "CASE WHEN (`a` = ?) THEN ? ELSE `b` > ? END",
			value:       []interface{}{1, 2, 3},
		},
		{
			name:        "case with nested when",
			caseBuilder: Case(When(Eq("col", "a"), 1), When(Eq("col", "b"), 2), Else(3)),
			query:       "CASE WHEN (`col` = ?) THEN ? WHEN (`col` = ?) THEN ? ELSE ? END",
			value:       []interface{}{"a", 1, "b", 2, 3},
		},
		{
			name:        "nested when with lt",
			caseBuilder: Case(When(Eq("colA", "a"), Lt("colB", 5)), When(Eq("colA", "b"), Lt("colB", 10)), Else(Lt("colB", 15))),
			query:       "CASE WHEN (`colA` = ?) THEN `colB` < ? WHEN (`colA` = ?) THEN `colB` < ? ELSE `colB` < ? END",
			value:       []interface{}{"a", 5, "b", 10, 15},
		},
		{
			name:        "nested when with lt as",
			caseBuilder: Case(When(Eq("colA", "a"), Lt("colB", 5)), When(Eq("colA", "b"), Lt("colB", 10)), Else(Lt("colB", 15))).As("c"),
			query:       "CASE WHEN (`colA` = ?) THEN `colB` < ? WHEN (`colA` = ?) THEN `colB` < ? ELSE `colB` < ? END AS `c`",
			value:       []interface{}{"a", 5, "b", 10, 15},
		},
	} {
		for _, sess := range testSession {
			t.Run(fmt.Sprintf("%s/%s", testSessionName(sess), test.name), func(t *testing.T) {
				reset(t, sess)

				expectedQuery := test.query
				if sess.Dialect != dialect.MySQL {
					// MySQL is the only one that uses a different QuoteIdent
					expectedQuery = strings.ReplaceAll(expectedQuery, "`", `"`)
				}

				buf := NewBuffer()
				err := test.caseBuilder.Build(sess.Dialect, buf)
				require.NoError(t, err)
				require.Equal(t, expectedQuery, buf.String())
				require.Equal(t, test.value, buf.Value())
			})
		}
	}
}

func TestSelectWithCase(t *testing.T) {
	for _, sess := range testSession {
		t.Run(testSessionName(sess), func(t *testing.T) {
			reset(t, sess)

			_, err := sess.InsertInto("dbr_people").
				Columns("name", "email").
				Values("test1", "test1@test.com").
				Values("test2", "test2@test.com").
				Values("test3", "test3@test.com").
				Exec()
			require.NoError(t, err)

			// select using case
			type people struct {
				Name  string
				Email string
				Case  int
			}
			var out []people
			stmt := sess.
				Select(
					"*",
					Case(
						When(Eq("name", "test2"), 100),
						Else(5),
					).As("case"),
				).
				From("dbr_people")
			count, err := stmt.Load(&out)

			require.NoError(t, err)
			require.Equal(t, 3, count)
			require.Equal(t, "test1", out[0].Name)
			require.Equal(t, "test1@test.com", out[0].Email)
			require.Equal(t, 5, out[0].Case)

			require.Equal(t, "test2", out[1].Name)
			require.Equal(t, "test2@test.com", out[1].Email)
			require.Equal(t, 100, out[1].Case)

			require.Equal(t, "test3", out[2].Name)
			require.Equal(t, "test3@test.com", out[2].Email)
			require.Equal(t, 5, out[2].Case)
		})
	}
}
