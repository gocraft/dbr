package dbr

import (
	"testing"

	"github.com/embrace-io/dbr/dialect"
	"github.com/stretchr/testify/require"
)

func TestComments(t *testing.T) {
	dialects := []Dialect{dialect.MySQL, dialect.PostgreSQL, dialect.SQLite3}
	for _, test := range []struct {
		name     string
		comments Comments
		expect   string
	}{
		{
			name:     "test comment",
			comments: Comments.Append(nil, "test comment").Append("another comment"),
			expect:   "/* test comment */\n/* another comment */\n",
		},
		{
			name:     "test comment with newline",
			comments: Comments.Append(nil, "test comment\nwith a newline").Append("another comment\nwith a newline"),
			expect:   "/* test comment\nwith a newline */\n/* another comment\nwith a newline */\n",
		},
		{
			name:     "test nested comment removed",
			comments: Comments.Append(nil, "/* test nested comment removed */"),
			expect:   "/* test nested comment removed */\n",
		},
	} {

		for _, d := range dialects {
			name := ""
			switch d {
			case dialect.MySQL:
				name = "MySQL"
			case dialect.PostgreSQL:
				name = "PostgreSQL"
			case dialect.SQLite3:
				name = "SQLite3"
			}
			t.Run(name+" "+test.name, func(t *testing.T) {
				buf := NewBuffer()
				test.comments.Build(d, buf)
				require.Equal(t, test.expect, buf.String())
			})
		}
	}

}
