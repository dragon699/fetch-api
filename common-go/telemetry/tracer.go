package telemetry

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)


type TracerConfig struct {
	ServiceName      string
	ServiceNamespace string
	ServiceVersion   string
	OtlpEndpointHTTP string
}

type TracerProvider struct {
	Tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metric.MeterProvider
}


func NewTracer(config TracerConfig) (*TracerProvider, error) {
	ctx := context.Background()
	endpoint := normalizeOTLPEndpoint(config.OtlpEndpointHTTP)

	traceExporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceNamespace(config.ServiceNamespace),
			semconv.ServiceVersion(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(10*time.Second))),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	tracer := otel.Tracer(config.ServiceName)

	return &TracerProvider{
		Tracer:         tracer,
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
	}, nil
}


func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	var err error

	if tp.meterProvider != nil {
		err = errors.Join(err, tp.meterProvider.Shutdown(ctx))
	}

	if tp.tracerProvider != nil {
		err = errors.Join(err, tp.tracerProvider.Shutdown(ctx))
	}

	return err
}

func normalizeOTLPEndpoint(endpoint string) string {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return trimmed
	}

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err == nil && parsed.Host != "" {
			return parsed.Host
		}
	}

	if strings.Contains(trimmed, "/") {
		return strings.Split(strings.Trim(trimmed, "/"), "/")[0]
	}

	return trimmed
}
