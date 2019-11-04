# LightStep OpenTelemetry Golang Exporter

This is an experimental exporter for opentelemetry-go.

## Initialize
```go
exporter := lightstep.NewExporter(lightstep.Config{
    accessToken: <PROJECT_ACCESS_TOKEN>,
    host: <SATELLITE_URL>,
    port: <SATELLITE_PORT>,
    serviceName: "my-service",
})

defer exporter.Close()
exporter.RegisterSimpleSpanProcessor()
```