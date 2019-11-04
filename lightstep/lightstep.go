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

// Config is a set of configuration options for LightStep.
type Config struct {
	// AccessToken is your LightStep project access token.
	// It can be found in the 'Settings' page for your project.
	AccessToken string
	// Host is the hostname for your LightStep Satellite(s).
	Host string
	// Port is the port number for your LightStep Satellite(s).
	Port int
	// ServiceName is an identifier for your application. This is displayed in the service directory.
	ServiceName string
}

// Exporter is an implementation of trace.Exporter that sends spans to LightStep.
type Exporter struct {
	once   sync.Once
	tracer ls.Tracer
}

func marshalConfigToOptions(c Config) ls.Options {
	opts := ls.Options{}
	opts.AccessToken = c.AccessToken
	opts.Collector.Host = c.Host
	opts.Collector.Port = c.Port
	opts.Collector.Plaintext = false

	return opts
}

// NewExporter is an implementation of trace.Exporter that sends spans to LightStep.
func NewExporter(config Config) (*Exporter, error) {
	tracerOptions := marshalConfigToOptions(config)
	tracer := ls.NewTracer(tracerOptions)

	opts := tracer.Options()
	if err := opts.Validate(); err != nil {
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

// Flush flushes all spans in the tracer.
// You should call this to flush spans to LightStep without closing the underlying tracer.
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
	first := binary.LittleEndian.Uint64(id[:8])
	second := binary.LittleEndian.Uint64(id[8:])
	return first ^ second
}

func convertSpanID(id core.SpanID) uint64 {
	return binary.LittleEndian.Uint64(id[:])
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
