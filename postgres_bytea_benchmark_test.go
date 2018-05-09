package dbr

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkByteaNoBinaryEncode(b *testing.B) {
	benchmarkBytea(b, postgresSession)
}

func BenchmarkByteaBinaryEncode(b *testing.B) {
	benchmarkBytea(b, postgresBinarySession)
}

func benchmarkBytea(b *testing.B, sess *Session) {
	data := bytes.Repeat([]byte("0123456789"), 1000)
	for _, v := range []string{
		`DROP TABLE IF EXISTS bytea_table`,
		`CREATE TABLE bytea_table (
			val bytea
		)`,
	} {
		_, err := sess.Exec(v)
		require.NoError(b, err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sess.InsertInto("bytea_table").Pair("val", data).Exec()
		require.NoError(b, err)
	}
}
