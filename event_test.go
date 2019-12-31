package dbr

import (
	"context"
)

type testTraceReceiver struct {
	NullEventReceiver
	started           []struct{ eventName, query string }
	errored, finished int
}

type testTraceReceiverWithContext struct {
	ctx context.Context
	err error
}

func (t *testTraceReceiverWithContext) EventWithContext(ctx context.Context, eventName string) {
	t.ctx = ctx
}

func (t *testTraceReceiverWithContext) EventKvWithContext(ctx context.Context, eventName string, kvs map[string]string) {
	t.ctx = ctx
}

func (t *testTraceReceiverWithContext) EventErrWithContext(ctx context.Context, eventName string, err error) error {
	t.ctx = ctx
	t.err = err
	return err
}

func (t *testTraceReceiverWithContext) EventErrKvWithContext(ctx context.Context, eventName string, err error, kvs map[string]string) error {
	t.ctx = ctx
	t.err = err
	return err
}

func (t *testTraceReceiverWithContext) TimingWithContext(ctx context.Context, eventName string, nanoseconds int64) {
	t.ctx = ctx
}

func (t *testTraceReceiverWithContext) TimingKvWithContext(ctx context.Context, eventName string, nanoseconds int64, kvs map[string]string) {
	t.ctx = ctx
}

func (t *testTraceReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	t.started = append(t.started, struct{ eventName, query string }{eventName, query})
	return ctx
}
func (t *testTraceReceiver) SpanError(ctx context.Context, err error) { t.errored++ }
func (t *testTraceReceiver) SpanFinish(ctx context.Context)           { t.finished++ }
