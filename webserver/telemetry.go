package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

var tracer = otel.Tracer("server")

func newTXTExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newGRPCExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	// Create the OTLP gRPC exporter pointing to Tempo
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("192.168.1.45:4317"),
		// otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())), // Tempo usually doesn't require TLS by default
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("goserver"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

var GlobalInstanceID = uuid.New()

func configureLogging() {
	// Use instance ID as prefix (helps in deployments with multiple instances)
	var (
		logger *slog.Logger
	)

	level := "debug"
	// Setup Default Logger
	if level != "debug" {
		// Production Logging
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})).
			With("server_id", GlobalInstanceID)
	} else {
		// Debug Logging
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})).
			With("server_id", GlobalInstanceID)
	}

	slog.SetDefault(logger)
	slog.Debug("Debug logging enabled")
}
