package dbr

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
)

//
// Your app can use these Null types instead of the defaults. The sole benefit you get is a MarshalJSON method that is not retarded.
//

type NullString struct {
	sql.NullString
}

type NullInt64 struct {
	sql.NullInt64
}

type NullTime struct {
	mysql.NullTime
}

type NullBool struct {
	sql.NullBool
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

func (n *NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Bool)
		return j, e
	}
	return nullString, nil
}
