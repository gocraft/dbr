package receivers

import (
	"github.com/gocraft/dbr"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type LoggingReceiverOpt interface {
	Apply(r *LoggingReceiver)
}

type loggingReceiverOptFunc func(r *LoggingReceiver)

func (f loggingReceiverOptFunc) Apply(r *LoggingReceiver) {
	f(r)
}

func Echo(enabled bool) LoggingReceiverOpt {
	return loggingReceiverOptFunc(func(r *LoggingReceiver) {
		r.echo = enabled
	})
}

func IncludeNotFound() LoggingReceiverOpt {
	return loggingReceiverOptFunc(func(r *LoggingReceiver) {
		r.includeNotFound = true
	})
}

type LoggingReceiver struct {
	logger *zap.Logger
	echo   bool
	includeNotFound bool
}

func NewLogReceiver(l *zap.Logger, opts ...LoggingReceiverOpt) *LoggingReceiver {
	r := &LoggingReceiver{logger: l}
	for _, opt := range opts {
		opt.Apply(r)
	}
	return r
}

func (l *LoggingReceiver) Event(eventName string) {
	l.logger.Info("received event", zap.String("event", eventName))
}

func (l *LoggingReceiver) EventKv(eventName string, kvs map[string]string) {
	fields := make([]zapcore.Field, 0, 10)
	fields = append(fields, zap.String("event", eventName))
	fields = append(fields, appendMapFields(fields, kvs)...)
	for key, value := range kvs {
		fields = append(fields, zap.String(key, value))
	}
	l.logger.Info("received event", fields...)
}

func (l *LoggingReceiver) EventErr(eventName string, err error) error {
	if !l.includeNotFound && err == dbr.ErrNotFound {
		return nil
	}
	l.logger.Error("received error event", zap.String("event", eventName), zap.Error(err))
	return nil
}

func (l *LoggingReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	if !l.includeNotFound && err == dbr.ErrNotFound {
		return nil
	}
	fields := make([]zapcore.Field, 0, 10)
	fields = append(fields, zap.String("event", eventName), zap.Error(err))
	fields = append(fields, appendMapFields(fields, kvs)...)
	l.logger.Error("received error event", fields...)
	return nil
}

func (l *LoggingReceiver) Timing(eventName string, nanoseconds int64) {
	if !l.echo {
		return
	}
	l.logger.Info("time elapsed for", zap.String("event", eventName), zap.Duration("elapsed", time.Duration(nanoseconds)))
}

func (l *LoggingReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	if !l.echo {
		return
	}
	fields := make([]zapcore.Field, 0, 10)
	fields = append(fields, zap.String("event", eventName))
	fields = append(fields, appendMapFields(fields, kvs)...)
	l.logger.Info("time elapsed for", fields...)
}

func appendMapFields(fields []zapcore.Field, kvs map[string]string) []zapcore.Field {
	fields = append(fields, zap.Namespace("dbr"))
	for key, value := range kvs {
		fields = append(fields, zap.String(key, value))
	}
	return fields
}

type MultiReceiver struct {
	receivers []dbr.EventReceiver
}

func (mr MultiReceiver) Event(eventName string) {
	for _, r := range mr.receivers {
		r.Event(eventName)
	}
}

func (mr MultiReceiver) EventKv(eventName string, kvs map[string]string) {
	for _, r := range mr.receivers {
		r.EventKv(eventName, kvs)
	}
}

func (mr MultiReceiver) EventErr(eventName string, err error) error {
	var rspErr error
	for _, r := range mr.receivers {
		rspErr = multierr.Append(err, r.EventErr(eventName, err))
	}
	return rspErr
}

func (mr MultiReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	var rspErr error
	for _, r := range mr.receivers {
		rspErr = multierr.Append(err, r.EventErrKv(eventName, err, kvs))
	}
	return rspErr
}

func (mr MultiReceiver) Timing(eventName string, nanoseconds int64) {
	for _, r := range mr.receivers {
		r.Timing(eventName, nanoseconds)
	}
}

func (mr MultiReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	for _, r := range mr.receivers {
		r.TimingKv(eventName, nanoseconds, kvs)
	}
}

func NewMultiReceiver(receivers ...dbr.EventReceiver) MultiReceiver {
	return MultiReceiver{receivers: receivers}
}

