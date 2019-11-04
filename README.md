# LightStep OpenTelemetry Golang Exporter

This is an experimental exporter for opentelemetry-go.

## Initialize
```go
exporter := lightstep.NewExporter([]lightstep.Option{
    lightstep.WithAccessToken(<PROJECT_ACCESS_TOKEN>),
    lightstep.WithHost(<SATELLITE_URL>),
    lightstep.WithPort(<SATELLITE_PORT>),
    lightstep.WithServiceName("my-service"),
})

defer exporter.Close()
exporter.RegisterSimpleSpanProcessor()
```
