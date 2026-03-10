package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

func Setup() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func FromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		logger = logger.With("trace_id", span.SpanContext().TraceID().String())
	}
	return logger
}
