package dialect

import (
	"time"
)

const (
	clickhouseTimeFormat = "2006-01-02 15:04:05"
)

type clickhouse struct {
	mysql
}

func (d clickhouse) EncodeTime(t time.Time) string {
	return `'` + t.UTC().Format(clickhouseTimeFormat) + `'`
}

func (d clickhouse) UpdateStmts() (string, string) {
	return "ALTER TABLE", "UPDATE"
}
