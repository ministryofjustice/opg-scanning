package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/felixge/httpsnoop"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func New(environment string) *slog.Logger {
	return slog.New(newSlogHandler(slog.NewJSONHandler(os.Stdout, nil)).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-scanning-service"),
			slog.String("environment", environment),
		}))
}

// Wraps the opg-go-common/telemetry packages StartTracerProvider.
func StartTracerProvider(ctx context.Context, logger *slog.Logger, exportTraces bool) (func(), error) {
	return telemetry.StartTracerProvider(ctx, logger, exportTraces)
}

func UseTelemetry(next http.Handler) http.HandlerFunc {
	tracer := otel.GetTracerProvider().Tracer("scanning")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.HTTPTargetKey.String(r.URL.Path)),
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("scanning", r.URL.Path, r)...),
		)
		defer span.End()

		m := httpsnoop.CaptureMetrics(next, w, r.WithContext(contextWithRequest(ctx, r)))

		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(m.Code)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(m.Code, trace.SpanKindServer))
		span.SetAttributes(semconv.HTTPResponseContentLengthKey.Int64(m.Written))
	}
}
