package dbr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComments(t *testing.T) {
	for _, test := range []struct {
		comments Comments
		expect   string
	}{
		{
			comments: Comments.Append(nil, "test comment").Append("another comment"),
			expect:   "/* test comment */\n/* another comment */\n",
		},
		{
			comments: Comments.Append(nil, "test comment\nwith a newline").Append("another comment\nwith a newline"),
			expect:   "/* test comment\nwith a newline */\n/* another comment\nwith a newline */\n",
		},
		{
			comments: Comments.Append(nil, "test comment\nwith a newline").Append("another comment\nwith a newline"),
			expect:   "/* test comment\nwith a newline */\n/* another comment\nwith a newline */\n",
		},
		{
			comments: Comments.Append(nil, "/* test nested comment removed */"),
			expect:   "/* test nested comment removed */\n",
		},
	} {
		buf := NewBuffer()
		test.comments.Write(buf)
		require.Equal(t, test.expect, buf.String())
	}

}
