package dbr

import (
	"log"
	"os"
	"time"
)

// EventReceiver gets events from dbr methods for logging purposes.
type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
	EventErr(eventName string, err error) error
	EventErrKv(eventName string, err error, kvs map[string]string) error
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

type kvs map[string]string

var nullReceiver = &NullEventReceiver{}

// NullEventReceiver is a sentinel EventReceiver.
// Use it if the caller doesn't supply one.
type NullEventReceiver struct{}

// Event receives a simple notification when various events occur.
func (n *NullEventReceiver) Event(eventName string) {}

// EventKv receives a notification when various events occur along with
// optional key/value data.
func (n *NullEventReceiver) EventKv(eventName string, kvs map[string]string) {}

// EventErr receives a notification of an error if one occurs.
func (n *NullEventReceiver) EventErr(eventName string, err error) error { return err }

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data.
func (n *NullEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	return err
}

// Timing receives the time an event took to happen.
func (n *NullEventReceiver) Timing(eventName string, nanoseconds int64) {}

// TimingKv receives the time an event took to happen along with optional key/value data.
func (n *NullEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {}

// PrintEventReceiver writes to anything that implements Printer.
// For example a *log.Logger
type PrintEventReceiver struct {
	Printer
}

// Printer interface matches log.Print and implementations should behave in a compatible manner.
type Printer interface {
	Print(v ...interface{})
}

// PrinterFunc implements Printer as a function.
type PrinterFunc func(v ...interface{})

func (f PrinterFunc) Print(v ...interface{}) { f(v...) }

// NewPrintEventReceiver creates an instance that prints to the Printer you provide.
// Passing nil will use a log.Logger that writes to os.Stderr.
func NewPrintEventReceiver(p Printer) *PrintEventReceiver {
	if p == nil {
		p = log.New(os.Stderr, "", log.LstdFlags)
	}
	return &PrintEventReceiver{
		Printer: p,
	}
}

// Event receives a simple notification when various events occur.
func (r *PrintEventReceiver) Event(eventName string) {
	r.Print(eventName)
}

// EventKv receives a notification when various events occur along with
// optional key/value data.
func (r *PrintEventReceiver) EventKv(eventName string, kvs map[string]string) {
	r.Print(eventName, ": ", kvs)
}

// EventErr receives a notification of an error if one occurs.
func (r *PrintEventReceiver) EventErr(eventName string, err error) error {
	r.Print(eventName, ", err: ", err)
	return err
}

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data.
func (r *PrintEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	r.Print(eventName, ": ", kvs, ", err: ", err)
	return err
}

// Timing receives the time an event took to happen.
func (r *PrintEventReceiver) Timing(eventName string, nanoseconds int64) {
	r.Print(eventName, ": timing: ", time.Duration(nanoseconds))
}

// TimingKv receives the time an event took to happen along with optional key/value data.
func (r *PrintEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	r.Print(eventName, ": ", kvs, ": timing: ", time.Duration(nanoseconds))

}
