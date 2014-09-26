package dbr

type expr struct {
	Sql    string
	Values []interface{}
}

// Expr is a SQL fragment with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *expr {
	return &expr{Sql: sql, Values: values}
}
