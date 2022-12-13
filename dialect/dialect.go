package dialect

import (
	"strings"
	"time"
)

var (
	// MySQL dialect
	MySQL = mysql{}
	// PostgreSQL dialect
	PostgreSQL = postgreSQL{}
	// SQLite3 dialect
	SQLite3 = sqlite3{}
	// MSSQL dialect
	MSSQL = mssql{}
)

const (
	timeFormat = time.RFC3339Nano
)

func quoteIdent(s, quote string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return quoteIdent(part[0], quote) + "." + quoteIdent(part[1], quote)
	}
	return quote + s + quote
}
