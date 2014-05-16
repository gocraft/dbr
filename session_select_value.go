package dbr

import (
	"fmt"
	"time"
)

func (sess *Session) SelectValue(dest interface{}, sql string, params ...interface{}) (bool, error) {
	// TODO: make sure dest is a ptr to something

	//
	// Get full SQL
	//
	fullSql, err := Interpolate(sql, params)
	if err != nil {
		return false, err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() {
		sess.TimingKv("dbr.select_value", time.Since(startTime).Nanoseconds(), map[string]string{"sql": fullSql})
	}()

	// Run the query:
	rows, err := sess.cxn.Db.Query(fullSql)
	if err != nil {
		fmt.Println("dbr.error.query") // Kvs{"error": err.String(), "sql": fullSql}
		return false, err
	}

	if rows.Next() {
		err = rows.Scan(dest)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (sess *Session) SelectUint64(sql string, params ...interface{}) (uint64, error) {
	var val uint64
	_, err := sess.SelectValue(&val, sql, params...)
	return val, err
}
