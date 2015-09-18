package dbr

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

//
// Your app can use these Null types instead of the defaults. The sole benefit you get is a MarshalJSON method that is not retarded.
//

// NullString is a type that can be null or a string
type NullString struct {
	sql.NullString
}

// NullFloat64 is a type that can be null or a float64
type NullFloat64 struct {
	sql.NullFloat64
}

// NullInt64 is a type that can be null or an int
type NullInt64 struct {
	sql.NullInt64
}

// NullTime is a type that can be null or a time
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	n.Time, n.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

// NullBool is a type that can be null or a bool
type NullBool struct {
	sql.NullBool
}

var nullString = []byte("null")

// MarshalJSON correctly serializes a NullString to JSON
func (n *NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullInt64 to JSON
func (n *NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int64)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullFloat64 to JSON
func (n *NullFloat64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Float64)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullTime to JSON
func (n *NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullBool to JSON
func (n *NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Bool)
	}
	return nullString, nil
}

// UnmarshalJSON correctly deserializes a NullString from JSON
func (n *NullString) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullInt64 from JSON
func (n *NullInt64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullFloat64 from JSON
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullTime from JSON
func (n *NullTime) UnmarshalJSON(b []byte) error {
	// scan for null
	if bytes.Equal(b, nullString) {
		return n.Scan(nil)
	}
	// scan for JSON timestamp
	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	return n.Scan(t)
}

// UnmarshalJSON correctly deserializes a NullBool from JSON
func (n *NullBool) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}
