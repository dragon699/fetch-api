package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


type TracerConfig struct {
	ServiceName      string
	ServiceNamespace string
	ServiceVersion   string
	OtlpEndpointGrpc string
}

type TracerProvider struct {
	Tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
}


func NewTracer(config TracerConfig) (*TracerProvider, error) {
	ctx := context.Background()

	// Create OTLP exporter
	conn, err := grpc.NewClient(
		config.OtlpEndpointGrpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// Create resource with service information
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

	// Create tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer := otel.Tracer(config.ServiceName)

	return &TracerProvider{
		Tracer:         tracer,
		tracerProvider: tracerProvider,
	}, nil
}


func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.tracerProvider != nil {
		return tp.tracerProvider.Shutdown(ctx)
	}
	return nil
}
