package dbr

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
)

type NullString struct {
	sql.NullString
}

type NullInt64 struct {
	sql.NullInt64
}

type NullTime struct {
	mysql.NullTime
}

var nullString = []byte("null")

func (n *NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.String)
		return j, e
	}
	return nullString, nil
}

func (n *NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Int64)
		return j, e
	}
	return nullString, nil
}

func (n *NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Time)
		return j, e
	}
	return nullString, nil
}
