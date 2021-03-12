package dialect

import (
	"strings"
	"time"
)

var (
	// MySQL dialect
	MySQL Dialect = mysql{}
	// PostgreSQL dialect
	PostgreSQL Dialect = postgreSQL{}
	// SQLite3 dialect
	SQLite3 Dialect = sqlite3{}
	// MSSQL dialect
	MSSQL Dialect = mssql{}
)

const (
	timeFormat = "2006-01-02 15:04:05.000000"
)

// Dialect abstracts database driver differences in encoding
// types, and placeholders.
type Dialect interface {
	QuoteIdent(id string) string

	EncodeString(s string) string
	EncodeBool(b bool) string
	EncodeTime(t time.Time) string
	EncodeBytes(b []byte) string

	Placeholder(n int) string
}

func quoteIdent(s, quote string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return quoteIdent(part[0], quote) + "." + quoteIdent(part[1], quote)
	}
	return quote + s + quote
}
