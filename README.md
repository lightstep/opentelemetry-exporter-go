# LightStep OpenTelemetry Golang Exporter

This is an experimental exporter for opentelemetry-go.

## Initialize
```go
func main() {
	exporter, err := lightstep.NewExporter([]lightstep.Option{
		lightstep.WithAccessToken(<PROJECT_ACCESS_TOKEN>),
		lightstep.WithHost(<SATELLITE_URL>),
		lightstep.WithPort(<SATELLITE_PORT>),
		lightstep.WithServiceName("my-service"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize Lightstep exporter: %v", err)
	}
	defer exporter.Close()

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter))
	global.SetTraceProvider(tp)

	[...]
}
```
