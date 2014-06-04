package dbr

import (
	"bytes"
)

// Use Mysql quoting by default
var Quoter = MysqlQuoter{}

// Interface for driver-swappable quoting
type quoter interface {
	writeQuotedColumn()
}

// Mysql-specific quoting
type MysqlQuoter struct{}

func (q MysqlQuoter) writeQuotedColumn(column string, sql *bytes.Buffer) {
	sql.WriteRune('`')
	sql.WriteString(column)
	sql.WriteRune('`')
}
