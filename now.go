package dbr

import (
	"database/sql/driver"
	"time"
)

type nowSentinel struct{}

// Now is a value that serializes to the current time
var Now = nowSentinel{}
var timeFormat = "2006-01-02 15:04:05"

// Value implements a valuer for compatibility
func (n nowSentinel) Value() (driver.Value, error) {
	now := time.Now().UTC().Format(timeFormat)
	return now, nil
}
