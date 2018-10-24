package opentracing

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// EventReceiver provides an embeddable implementation of dbr.TracingEventReceiver
// powered by opentracing-go.
type EventReceiver struct{}

// SpanStart starts a new query span from ctx, then returns a new context with the new span.
func (EventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	span, ctx := ot.StartSpanFromContext(ctx, eventName)
	otext.DBStatement.Set(span, query)
	otext.DBType.Set(span, "sql")
	return ctx
}

// SpanFinish finishes the span associated with ctx.
func (EventReceiver) SpanFinish(ctx context.Context) {
	if span := ot.SpanFromContext(ctx); span != nil {
		span.Finish()
	}
}

// SpanError adds an error to the span associated with ctx.
func (EventReceiver) SpanError(ctx context.Context, err error) {
	if span := ot.SpanFromContext(ctx); span != nil {
		otext.Error.Set(span, true)
		span.LogFields(otlog.String("event", "error"), otlog.Error(err))
	}
}
