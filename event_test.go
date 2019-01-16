package dbr

import (
	"context"
)

type testTraceReceiver struct {
	NullEventReceiver
	started           []struct{ eventName, query string }
	errored, finished int
}

func (t *testTraceReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	t.started = append(t.started, struct{ eventName, query string }{eventName, query})
	return ctx
}
func (t *testTraceReceiver) SpanError(ctx context.Context, err error) { t.errored++ }
func (t *testTraceReceiver) SpanFinish(ctx context.Context)           { t.finished++ }
