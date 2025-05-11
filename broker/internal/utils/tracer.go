package utils

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"runtime"
)

var Tracer = otel.GetTracerProvider().Tracer("")

func StartSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	return Tracer.Start(ctx, details.Name())
}
