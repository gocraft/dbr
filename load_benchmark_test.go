package dbr

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func BenchmarkLoadValues(b *testing.B) {
	sess := mysqlSession
	for _, v := range []string{
		`DROP TABLE IF EXISTS suggestions`,
		`CREATE TABLE suggestions (
			id serial PRIMARY KEY,
			title varchar(255),
			body text
		)`,
	} {
		_, err := sess.Exec(v)
		require.NoError(b, err)
	}
	tx, err := sess.Begin()
	require.NoError(b, err)

	const maxRows = 100000

	for i := 0; i < maxRows; i++ {
		_, err := tx.InsertInto("suggestions").
			Columns("title", "body").
			Values("title", "body").
			Exec()
		require.NoError(b, err)
	}
	err = tx.Commit()
	require.NoError(b, err)

	type Suggestion struct {
		Title *string
		Body  *string
	}
	for n := 10; n <= maxRows; n *= 10 {
		query := fmt.Sprintf("SELECT * FROM suggestions ORDER BY id ASC LIMIT %d", n)

		b.Run(fmt.Sprintf("sqlx_%d", n), func(b *testing.B) {
			b.StopTimer()
			db, err := sqlx.Connect("mysql", mysqlDSN)
			require.NoError(b, err)
			db = db.Unsafe()
			defer db.Close()

			for i := 0; i < b.N; i++ {
				var suggs []*Suggestion
				b.StartTimer()
				err := db.SelectContext(context.Background(), &suggs, query)
				b.StopTimer()
				require.NoError(b, err)
				require.Len(b, suggs, n)
			}
		})
		b.Run(fmt.Sprintf("dbr_%d", n), func(b *testing.B) {
			b.StopTimer()

			for i := 0; i < b.N; i++ {
				var suggs []*Suggestion
				b.StartTimer()
				_, err := sess.SelectBySql(query).LoadContext(context.Background(), &suggs)
				b.StopTimer()
				require.NoError(b, err)
				require.Len(b, suggs, n)
			}
		})
	}
}
