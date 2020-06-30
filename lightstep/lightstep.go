package lightstep

import (
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/opentracing/opentracing-go"

	ls "github.com/lightstep/lightstep-tracer-go"
)

// Option struct is used to configre the LightStepExpoter options.
type Option func(*config)

// WithAccessToken sets the LightStep access token used to authenticate and associate data.
// with a LightStep project
func WithAccessToken(accessToken string) Option {
	return func(c *config) {
		c.options.AccessToken = accessToken
	}
}

// WithHost sets the URL hostname for sending data to LightStep satellites.
func WithHost(host string) Option {
	return func(c *config) {
		c.options.Collector.Host = host
		c.options.SystemMetrics.Endpoint.Host = host
	}
}

// WithPort sets the URL port for sending data to LightStep satellites.
func WithPort(port int) Option {
	return func(c *config) {
		c.options.Collector.Port = port
		c.options.SystemMetrics.Endpoint.Port = port
	}
}

// WithServiceName sets the service name tag used to identify a service in
// the LightStep application.
func WithServiceName(serviceName string) Option {
	return func(c *config) {
		if c.options.Tags == nil {
			c.options.Tags = make(map[string]interface{})
		}

		c.options.Tags[ls.ComponentNameKey] = serviceName
	}
}

// WithServiceVersion sets the service version used to identify a service's
// version in the LightStep application.
func WithServiceVersion(serviceVersion string) Option {
	return func(c *config) {
		if c.options.Tags == nil {
			c.options.Tags = make(map[string]interface{})
		}

		c.options.Tags[ls.ServiceVersionKey] = serviceVersion
	}
}

// WithPlainText indicates if data should be sent in plaintext to the LightStep
// Satelites. Default is false.
func WithPlainText(pt bool) Option {
	return func(c *config) {
		c.options.Collector.Plaintext = pt
		c.options.SystemMetrics.Endpoint.Plaintext = pt
	}
}

// WithSystemMetricsDisabled determines if system metrics are disabled or not.
// Default is false.
func WithSystemMetricsDisabled(disabled bool) Option {
	return func(c *config) {
		c.options.SystemMetrics.Disabled = disabled
	}
}

// WithSystemMetricTimeout sets the tineout duration for sending metrics
// reports to the LightStep application.
func WithSystemMetricTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.options.SystemMetrics.Timeout = timeout
	}
}

// WithSystemMetricMeasurementFrequency sets the tineout duration for sending metrics
// reports to the LightStep application.
func WithSystemMetricMeasurementFrequency(frequency time.Duration) Option {
	return func(c *config) {
		c.options.SystemMetrics.MeasurementFrequency = frequency
	}
}

type config struct {
	options ls.Options
}

func newConfig(opts ...Option) config {
	var c config
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	return c
}

// Exporter is an implementation of trace.Exporter that sends spans to LightStep.
type Exporter struct {
	once   sync.Once
	tracer ls.Tracer
}

// NewExporter is an implementation of trace.Exporter that sends spans to LightStep.
func NewExporter(opts ...Option) (*Exporter, error) {
	c := newConfig(opts...)
	tracer := ls.NewTracer(c.options)

	checkOptions := tracer.Options()
	if err := checkOptions.Validate(); err != nil {
		return nil, err
	}

	return &Exporter{
		tracer: tracer,
	}, nil
}

// ExportSpan exports an OpenTelementry SpanData object to an OpenTracing Span on the LightStep tracer.
func (e *Exporter) ExportSpan(ctx context.Context, data *trace.SpanData) {
	e.tracer.StartSpan(
		data.Name,
		ls.SetTraceID(convertTraceID(data.SpanContext.TraceID)),
		ls.SetSpanID(convertSpanID(data.SpanContext.SpanID)),
		ls.SetParentSpanID(convertSpanID(data.ParentSpanID)),
		opentracing.StartTime(data.StartTime),
		opentracing.Tags(toTags(data.Attributes)),
	).FinishWithOptions(
		opentracing.FinishOptions{
			FinishTime: data.EndTime,
			LogRecords: toLogRecords(data.MessageEvents),
		},
	)
}

var _ trace.SpanSyncer = (*Exporter)(nil)

// Close flushes all spans in the tracer to LightStep and then closes the tracer.
// You should call Close() before your application exits.
func (e *Exporter) Close() {
	e.tracer.Close(context.Background())
}

// Flush forces all unflushed to flush.
// This is normally handled by the exporter. However, you may call this to explicitly flush all spans without closing the exporter.
func (e *Exporter) Flush() {
	e.tracer.Flush(context.Background())
}

// this replicates StartSpan behavior for testing
func lightStepSpan(data *trace.SpanData) *ls.RawSpan {
	spanContext := ls.SpanContext{
		TraceID: convertTraceID(data.SpanContext.TraceID),
		SpanID:  convertSpanID(data.SpanContext.SpanID),
	}
	lsSpan := &ls.RawSpan{
		Context:      spanContext,
		ParentSpanID: convertSpanID(data.ParentSpanID),
		Operation:    data.Name,
		Start:        data.StartTime,
		Tags:         toTags(data.Attributes),
		Logs:         toLogRecords(data.MessageEvents),
	}
	lsSpan.Duration = data.EndTime.Sub(data.StartTime)
	return lsSpan
}

func convertTraceID(id core.TraceID) uint64 {
	return binary.BigEndian.Uint64(id[8:])
}

func convertSpanID(id core.SpanID) uint64 {
	return binary.BigEndian.Uint64(id[:])
}

func toLogRecords(input []trace.Event) []opentracing.LogRecord {
	output := make([]opentracing.LogRecord, 0, len(input))
	for _, l := range input {
		output = append(output, toLogRecord(l))
	}
	return output
}

func toTags(input []core.KeyValue) map[string]interface{} {
	output := make(map[string]interface{})
	for _, value := range input {
		output[string(value.Key)] = value.Value.AsInterface()
	}
	return output
}

func toLogRecord(ev trace.Event) opentracing.LogRecord {
	return opentracing.LogRecord{
		Timestamp: ev.Time,
		Fields:    toFields(ev.Attributes),
	}
}

func toFields(input []core.KeyValue) []log.Field {
	output := make([]log.Field, 0, len(input))
	for _, value := range input {
		output = append(output, log.Object(string(value.Key), value.Value.AsInterface()))
	}
	return output
}
