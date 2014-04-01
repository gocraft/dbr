package dbr

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

// target can be:
// - addr of a structure
// - addr of slice of pointers to structures
// - map of pointers to structures (addr of map also ok)
// If it's a single structure, only the first record returned will be set.
// If it's a slice or map, the slice/map won't be emptied first. New records will be allocated for each found record.
// If its a map, there is the potential to overwrite values (keys are 'id')
func (sess *Session) FindBySql(sql string, target interface{}) error {
	// validate target
	
	sess.cxn.Db.Query()
	
	return nil
}




