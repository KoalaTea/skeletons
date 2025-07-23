package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/koalatea/go-project-skeleton/ent/runtime"

	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel"
)

func main() {
	ctx := context.Background()

	f, err := os.Create("traces.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	exp, err := newTXTExporter(f)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	// tracer = tp.Tracer("exampleserver")

	server := newServer()
	fmt.Println("starting server")
	if err := server.Run(ctx); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}
