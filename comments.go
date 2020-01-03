package dbr

import "strings"

const (
	openingSQLComment = "/*"
	closingSQLComment = "*/"
	space             = " "
	newline           = "\n"
	emptyString       = ""
)

// Comments represents a set of sql comments
type Comments []string

// Append a new sql comment to a set of comments
func (comments Comments) Append(comment string) Comments {
	comment = strings.Replace(comment, openingSQLComment, emptyString, -1)
	comment = strings.Replace(comment, closingSQLComment, emptyString, -1)
	comment = strings.TrimSpace(comment)
	comments = append(comments, comment)
	return comments
}

// Build writes each comment in the form of "/* some comment */\n"
func (comments Comments) Build(d Dialect, buf Buffer) error {
	for _, comment := range comments {
		words := []string{openingSQLComment, space, comment, space, closingSQLComment, newline}
		for _, str := range words {
			if _, err := buf.WriteString(str); err != nil {
				return err
			}
		}
	}
	return nil
}
52 comments_test.go
@@ -0,0 +1,52 @@
package dbr

import (
	"testing"

	"github.com/gocraft/dbr/dialect"
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
