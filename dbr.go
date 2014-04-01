package dbr

import (
	"fmt"
)

type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
}

type TimingReceiver interface {
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

func DoSomething(s EventReceiver) {
	fmt.Println("Doing it.", s)
	
	s.Event("sup")
}