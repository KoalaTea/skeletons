package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/koalatea/authserver/server/ent/runtime"

	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

var tracer = otel.Tracer("authserver")

func newExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("authserver"),
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

func main() {
	ctx := context.Background()

	f, err := os.Create("traces.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	exp, err := newExporter(f)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	// tracer = tp.Tracer("authserver")

	server := newServer(ctx)
	fmt.Println("starting server")
	if err := server.Run(ctx); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}
