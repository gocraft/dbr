package dbr

import (
	"fmt"
)

// Use Mysql quoting by default
var Quoter = MysqlQuoter{}

// Interface for driver-swappable quoting
type quoter interface {
	QuoteColumn()
}

// Mysql-specific quoting
type MysqlQuoter struct{}

func (q MysqlQuoter) QuoteColumn(column string) string {
	return fmt.Sprintf("`%s`", column)
}
