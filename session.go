package dbr

import (
	"fmt"
	"reflect"
)


type Session struct {
	cxn *Connection
	log EventReceiver
}

func (cxn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = cxn.log
	}
	return &Session{cxn: cxn, log: log}
}

// Given a query and given a structure (field list), there's 2 sets of fields.
// Take the intersection. We can fill those in. great.
// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
// For fields in the query that aren't in the structure, we'll ignore them.

// dest can be:
// - addr of a structure
// - addr of slice of pointers to structures
// - map of pointers to structures (addr of map also ok)
// If it's a single structure, only the first record returned will be set.
// If it's a slice or map, the slice/map won't be emptied first. New records will be allocated for each found record.
// If its a map, there is the potential to overwrite values (keys are 'id')
// Returns the number of items found (which is not necessarily the # of items set)
func (sess *Session) FindBySql(dest interface{}, sql string, params ...interface{}) (int, error) {
	// validate target
	
	valueOfDest := reflect.ValueOf(dest)
	
	
	fullSql := sql
	
	numberOfRowsReturned := 0
	
	// Run the query:
	rows, err := sess.cxn.Db.Query(fullSql)
	if err != nil {
		fmt.Println("dbr.error.query") // Kvs{"error": err.String(), "sql": fullSql}
		return 0, err
	}
	
	// Get the columns:
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("dbr.error.columns")
		return 0, err
	}
	fmt.Println("cols = ", columns)
	
	// Given columns, and given the dest record type (eg, Suggsetion)
	fieldMap
	
	var holder []interface{}
	
	holder = make([]interface{}, 0, len(columns))
	
	
	
	for rows.Next() {
		
		
		
        err := rows.Scan(holder...)
        if err != nil {
          return count, err
        }
		
		numberOfRowsReturned += 1
	}
	
    // Check for errors at the end. Supposedly these are error that can happen during iteration.
    if err = rows.Err(); err != nil {
      return numberOfRowsReturned, err
    }
	
	return numberOfRowsReturned, nil
}


