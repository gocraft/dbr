package dialect

import (
	"fmt"
	"time"
)

type postgreSQL struct{}

func (d postgreSQL) QuoteIdent(s string) string {
	return quoteIdent(s, `"`)
}

func (d postgreSQL) EncodeString(s string) string {
	return MySQL.EncodeString(s)
}

func (d postgreSQL) EncodeBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func (d postgreSQL) EncodeTime(t time.Time) string {
	return MySQL.EncodeTime(t)
}

func (d postgreSQL) EncodeBytes(b []byte) string {
	return d.EncodeString(fmt.Sprintf(`\x%x`, b))
}

func (d postgreSQL) Placeholder() string {
	return "?"
}
