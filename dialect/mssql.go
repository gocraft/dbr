package dialect

import (
	"fmt"
	"strings"
	"time"
)

type mssql struct{}

func (d mssql) QuoteIdent(s string) string {
	return quoteIdent(s, `"`)
}

func (d mssql) EncodeString(s string) string {
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func (d mssql) EncodeBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (d mssql) EncodeTime(t time.Time) string {
	return t.Format("'2006-01-02 15:04:05.999'")
}

func (d mssql) EncodeBytes(b []byte) string {
	return fmt.Sprintf(`E'\\x%x'`, b)
}

func (d mssql) Placeholder(n int) string {
	return fmt.Sprintf("@p%d", n+1)
}
