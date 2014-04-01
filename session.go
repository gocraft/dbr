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

