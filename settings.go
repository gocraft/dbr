package dbr

import (
	"fmt"
	"strings"

	"github.com/gocraft/dbr/v2/dialect"
)

type QuerySettings []string

func (qs QuerySettings) Append(setting, value string) QuerySettings {
	setting = strings.TrimSpace(setting)
	value = strings.TrimSpace(value)
	qs = append(qs, fmt.Sprintf("%s = %s", setting, value))
	return qs
}

// Build writes each setting in the form of "SETTINGS setting_key=value \n"
func (qs QuerySettings) Build(d Dialect, buf Buffer) error {
	// Only clickhouse supports settings, so we don't build anything if it's a different dialect.
	if d != dialect.ClickHouse {
		return nil
	}
	for _, setting := range qs {
		if _, err := buf.WriteString("\nSETTINGS " + setting); err != nil {
			return err
		}
	}
	return nil
}
