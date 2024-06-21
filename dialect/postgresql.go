package dialect

import (
	"fmt"
	"strings"
	"time"
)

type postgreSQL struct{}

func (d postgreSQL) QuoteIdent(s string) string {
	return quoteIdent(s, `"`)
}

func (d postgreSQL) EncodeString(s string) string {
	// http://www.postgresql.org/docs/9.2/static/sql-syntax-lexical.html
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func (d postgreSQL) EncodeBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func (d postgreSQL) EncodeTime(t time.Time) string {
	return `'` + t.Format(time.RFC3339Nano) + `'`
}

func (d postgreSQL) EncodeBytes(b []byte) string {
	return fmt.Sprintf(`E'\\x%x'`, b)
}

func (d postgreSQL) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n+1)
}

func (d postgreSQL) UpdateStmts() (string, string) {
	return "UPDATE", "SET"
}

func (d postgreSQL) OnConflict(constraint string) string {
	return fmt.Sprintf("ON CONFLICT ON CONSTRAINT %s DO UPDATE SET", d.QuoteIdent(constraint))
}

func (d postgreSQL) Proposed(column string) string {
	return fmt.Sprintf("EXCLUDED.%s", d.QuoteIdent(column))
}
