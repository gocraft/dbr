package dbr

import (
	"testing"

	"github.com/embrace-io/dbr/dialect"
	"github.com/stretchr/testify/require"
)

func TestQuerySettings(t *testing.T) {
	dialects := []Dialect{dialect.MySQL, dialect.PostgreSQL, dialect.SQLite3, dialect.Clickhouse}
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
		for _, d := range dialects {
			name := ""
			switch d {
			case dialect.MySQL:
				name = "MySQL"
			case dialect.PostgreSQL:
				name = "PostgreSQL"
			case dialect.SQLite3:
				name = "SQLite3"
			case dialect.Clickhouse:
				name = "ClickHouse"
			}
			t.Run(name+" "+test.name, func(t *testing.T) {
				buf := NewBuffer()
				test.settings.Build(d, buf)
				actual := buf.String()
				if d == dialect.Clickhouse {
					require.Equal(t, test.expect, actual)
				} else {
					require.Equal(t, "", actual)
				}
			})
		}
	}

}
