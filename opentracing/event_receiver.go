package opentracing

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type EventReceiver struct{}

func (EventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	span, ctx := ot.StartSpanFromContext(ctx, eventName)
	otext.DBStatement.Set(span, query)
	otext.DBType.Set(span, "sql")
	return ctx
}

func (EventReceiver) SpanFinish(ctx context.Context) {
	if span := ot.SpanFromContext(ctx); span != nil {
		span.Finish()
	}
}

func (EventReceiver) SpanError(ctx context.Context, err error) {
	if span := ot.SpanFromContext(ctx); span != nil {
		otext.Error.Set(span, true)
		span.LogFields(otlog.String("event", "error"), otlog.Error(err))
	}
}
