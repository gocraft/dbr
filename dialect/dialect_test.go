package dialect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMySQL(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{
			in:   "table.col",
			want: "`table`.`col`",
		},
		{
			in:   "col",
			want: "`col`",
		},
	} {
		assert.Equal(t, test.want, MySQL.QuoteIdent(test.in))
	}
}

func TestPostgreSQL(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{
			in:   "table.col",
			want: `"table"."col"`,
		},
		{
			in:   "col",
			want: `"col"`,
		},
	} {
		assert.Equal(t, test.want, PostgreSQL.QuoteIdent(test.in))
	}
}

func TestSQLite3(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{
			in:   "table.col",
			want: `"table"."col"`,
		},
		{
			in:   "col",
			want: `"col"`,
		},
	} {
		assert.Equal(t, test.want, SQLite3.QuoteIdent(test.in))
	}
}
