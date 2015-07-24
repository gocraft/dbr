package dbr

import (
  "bytes"
)

// Quoter is the quoter to use for quoting text; use Mysql quoting by default
var Quoter quoter = MysqlQuoter{}

// Interface for driver-swappable quoting
type quoter interface {
  writeQuotedColumn(column string, sql *bytes.Buffer)
}

// MysqlQuoter implements Mysql-specific quoting
type MysqlQuoter struct{}

func (q MysqlQuoter) writeQuotedColumn(column string, sql *bytes.Buffer) {
  sql.WriteRune('`')
  sql.WriteString(column)
  sql.WriteRune('`')
}

// PgQuoter implements Postgres-specific quoting
type PgQuoter struct{}

func (q PgQuoter) writeQuotedColumn(column string, sql *bytes.Buffer) {
  sql.WriteRune('"')
  sql.WriteString(column)
  sql.WriteRune('"')
}

func SetQuoter(q quoter) {
  Quoter = q
}
