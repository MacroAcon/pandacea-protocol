//go:build otel

package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlpmetrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init configures OpenTelemetry exporters and providers.
func Init(ctx context.Context, logger *slog.Logger) (func(context.Context) error, error) {
	serviceName := "agent-backend"
	env := os.Getenv("DEPLOYMENT_ENV")
	if env == "" {
		env = os.Getenv("PANDACEA_ENV")
		if env == "" {
			env = "development"
		}
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("deployment.environment", env),
		),
		sdkresource.WithFromEnv(),
	)
	if err != nil {
		return nil, err
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}

	// Traces
	traceExp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Metrics
	metricExp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	if err != nil {
		logger.Warn("metrics exporter init failed", "error", err)
	}
	var mp *sdkmetric.MeterProvider
	if metricExp != nil {
		reader := sdkmetric.NewPeriodicReader(metricExp)
		mp = sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
		otel.SetMeterProvider(mp)
	}

	return func(ctx context.Context) error {
		var merr error
		if mp != nil {
			if err := mp.Shutdown(ctx); err != nil {
				merr = err
			}
		}
		if err := tp.Shutdown(ctx); err != nil {
			if merr == nil {
				merr = err
			}
		}
		return merr
	}, nil
}
