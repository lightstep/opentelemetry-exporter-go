# Lightstep OpenTelemetry Golang Exporter

This is a Lightstep exporter for opentelemetry-go.

## Initialize

This example connects to Lightstep and sends a single span.

```go
package main

import (
	"context"
	"log"

	"github.com/lightstep/opentelemetry-exporter-go/lightstep"
	"go.opentelemetry.io/otel/api/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	exporter, err := lightstep.NewExporter(
		lightstep.WithAccessToken("<ACCESS_TOKEN>"),
		lightstep.WithServiceName("my-service"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Lightstep exporter: %v", err)
	}
	defer exporter.Close()

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter))
	global.SetTraceProvider(tp)

	ctx := context.Background()
	_, span := global.Tracer("example").Start(ctx, "hello")
	span.End()

	exporter.Flush()
}
```
