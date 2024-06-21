package dbr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComments(t *testing.T) {
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
		for _, sess := range testSession {
			t.Run(fmt.Sprintf("%s/%s", testSessionName(sess), test.name), func(t *testing.T) {
				buf := NewBuffer()
				err := test.comments.Build(sess.Dialect, buf)
				require.NoError(t, err)
				require.Equal(t, test.expect, buf.String())

				stmt := sess.Select("1")
				stmt.comments = test.comments

				buf2 := NewBuffer()
				err = stmt.Build(sess.Dialect, buf2)
				require.NoError(t, err)
				require.Equal(t, test.expect+"SELECT 1", buf2.String())

				one, err := stmt.ReturnInt64()
				require.NoError(t, err)
				require.EqualValues(t, 1, one)
			})
		}
	}
}
