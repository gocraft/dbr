package dbr

import (
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		j, e := json.Marshal(ns.String)
		return j, e
	}
	return []byte("null"), nil
}