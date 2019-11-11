package lightstep

import (
	"context"
	"encoding/binary"
	"sync"

	"github.com/opentracing/opentracing-go/log"

	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/sdk/export"

	"github.com/opentracing/opentracing-go"

	ls "github.com/lightstep/lightstep-tracer-go"
)

type ExporterOption func(*config)

func WithAccessToken(accessToken string) ExporterOption {
	return func(c *config) {
		c.options.AccessToken = accessToken
	}
}

func WithHost(host string) ExporterOption {
	return func(c *config) {
		c.options.Collector.Host = host
	}
}

func WithPort(port int) ExporterOption {
	return func(c *config) {
		c.options.Collector.Port = port
	}
}

func WithServiceName(serviceName string) ExporterOption {
	return func(c *config) {
		if c.options.Tags == nil {
			c.options.Tags = make(map[string]interface{})
		}

		c.options.Tags[ls.ComponentNameKey] = serviceName
	}
}

type config struct {
	options ls.Options
}

func newConfig(opts ...ExporterOption) config {
	var c config
	var defaultOpts []ExporterOption

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
func NewExporter(opts ...ExporterOption) (*Exporter, error) {
	c := newConfig(opts...)
	tracer := ls.NewTracer(c.options)

	tracerOptions := tracer.Options()
	if err := tracerOptions.Validate(); err != nil {
		return nil, err
	}

	return &Exporter{
		tracer: tracer,
	}, nil
}

// ExportSpan exports an OpenTelementry SpanData object to an OpenTracing Span on the LightStep tracer.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
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

var _ export.SpanSyncer = (*Exporter)(nil)

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
func lightStepSpan(data *export.SpanData) *ls.RawSpan {
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
	return binary.BigEndian.Uint64(id[:8])
}

func convertSpanID(id core.SpanID) uint64 {
	return binary.BigEndian.Uint64(id[:])
}

func toLogRecords(input []export.Event) []opentracing.LogRecord {
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

func toLogRecord(ev export.Event) opentracing.LogRecord {
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
