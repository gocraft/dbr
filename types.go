package dbr

import (
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}

type NullInt64 struct {
	sql.NullInt64
}

func (n *NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.String)
		return j, e
	}
	return []byte("null"), nil
}

func (n *NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Int64)
		return j, e
	}
	return []byte("null"), nil
}