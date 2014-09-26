package dbr

import (
	"bytes"
)

// Quoter is the quoter to use for quoting text; use Mysql quoting by default
var Quoter = MysqlQuoter{}

// Interface for driver-swappable quoting
type quoter interface {
	writeQuotedColumn()
}

// MysqlQuoter implements Mysql-specific quoting
type MysqlQuoter struct{}

func (q MysqlQuoter) writeQuotedColumn(column string, sql *bytes.Buffer) {
	sql.WriteRune('`')
	sql.WriteString(column)
	sql.WriteRune('`')
}
