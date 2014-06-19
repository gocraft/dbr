package dbr

type expr struct {
	Sql    string
	Values []interface{}
}

func Expr(sql string, values ...interface{}) *expr {
	return &expr{Sql: sql, Values: values}
}
