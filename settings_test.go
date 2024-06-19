package dbr

import (
	"fmt"
	"testing"

	"github.com/gocraft/dbr/v2/dialect"
	"github.com/stretchr/testify/require"
)

func TestQuerySettings(t *testing.T) {
	for _, test := range []struct {
		name     string
		settings QuerySettings
		expect   string
	}{
		{
			name:     "test single setting",
			settings: QuerySettings.Append(nil, "key", "value"),
			expect:   "\nSETTINGS key = value",
		},
		{
			name:     "test multiple setting",
			settings: QuerySettings.Append(nil, "key1", "value1").Append("key2", "value2"),
			expect:   "\nSETTINGS key1 = value1\nSETTINGS key2 = value2",
		},
		{
			name:     "test trimming setting args",
			settings: QuerySettings.Append(nil, " key_needs_trimming   ", " value_needs_trimming    "),
			expect:   "\nSETTINGS key_needs_trimming = value_needs_trimming",
		},
	} {
		for _, sess := range testSession {
			t.Run(fmt.Sprintf("%s/%s", testSessionName(sess), test.name), func(t *testing.T) {
				buf := NewBuffer()
				err := test.settings.Build(sess.Dialect, buf)
				require.NoError(t, err)

				actual := buf.String()
				if sess.Dialect == dialect.Clickhouse {
					require.Equal(t, test.expect, actual)
				} else {
					require.Equal(t, "", actual)
				}
			})
		}
	}
}
