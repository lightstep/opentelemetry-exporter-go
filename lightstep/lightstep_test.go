package lightstep

import (
	"testing"
	"time"

	ls "github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/api/kv"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestExport(t *testing.T) {
	assert := assert.New(t)
	now := time.Now().Round(time.Microsecond)

	traceID, _ := apitrace.IDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := apitrace.SpanIDFromHex("0102030405060708")

	expectedTraceID := uint64(0x90a0b0c0d0e0f10)
	expectedSpanID := uint64(0x102030405060708)

	tests := []struct {
		name string
		data *trace.SpanData
		want *ls.RawSpan
	}{
		{
			name: "root span",
			data: &trace.SpanData{
				SpanContext: apitrace.SpanContext{
					TraceID: traceID,
					SpanID:  spanID,
				},
				Name:      "/test",
				StartTime: now,
				EndTime:   now,
				Resource: resource.New(
					kv.String("R1", "V1"),
				),
				Attributes: []kv.KeyValue{
					kv.String("A", "B"),
					kv.String("C", "D"),
				},
			},
			want: &ls.RawSpan{
				Context: ls.SpanContext{
					TraceID: expectedTraceID,
					SpanID:  expectedSpanID,
				},
				Operation: "/test",
				Start:     now,
				Duration:  0,
				Tags: opentracing.Tags{
					"A":  "B",
					"C":  "D",
					"R1": "V1",
				},
			},
		},
		{
			name: "with events",
			data: &trace.SpanData{
				SpanContext: apitrace.SpanContext{
					TraceID: traceID,
					SpanID:  spanID,
				},
				Name:      "/test",
				StartTime: now,
				EndTime:   now,
				MessageEvents: []trace.Event{
					trace.Event{
						Name: "myevent",
						Attributes: []kv.KeyValue{
							kv.String("A", "B"),
						},
						Time: now,
					},
				},
				Resource: resource.New(
					kv.String("R1", "V1"),
				),
			},
			want: &ls.RawSpan{
				Context: ls.SpanContext{
					TraceID: expectedTraceID,
					SpanID:  expectedSpanID,
				},
				Operation: "/test",
				Start:     now,
				Duration:  0,
				Tags: opentracing.Tags{
					"R1": "V1",
				},
				Logs: []opentracing.LogRecord{
					opentracing.LogRecord{
						Timestamp: now,
						Fields: []log.Field{
							log.String("name", "myevent"),
							log.Object("A", "B"),
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		lsSpan := lightStepSpan(test.data)
		assert.EqualValues(test.want, lsSpan)
	}
}

func TestWithServiceVersion(t *testing.T) {
	assert := assert.New(t)

	serviceVersion := "1.0.0"
	config := newConfig(
		WithServiceVersion(serviceVersion),
	)

	assert.EqualValues(serviceVersion, config.options.Tags[ls.ServiceVersionKey])
}

func TestWithHost(t *testing.T) {
	assert := assert.New(t)

	host := "my.host.com"
	config := newConfig(
		WithHost(host),
	)

	assert.EqualValues(host, config.options.Collector.Host)
	assert.EqualValues(host, config.options.SystemMetrics.Endpoint.Host)
}

func TestWithPort(t *testing.T) {
	assert := assert.New(t)

	port := 123
	config := newConfig(
		WithPort(port),
	)

	assert.EqualValues(port, config.options.Collector.Port)
	assert.EqualValues(port, config.options.SystemMetrics.Endpoint.Port)
}

func TestWithAccessToken(t *testing.T) {
	assert := assert.New(t)

	token := "my-token"
	config := newConfig(
		WithAccessToken(token),
	)

	assert.EqualValues(token, config.options.AccessToken)
}

func TestWithServiceName(t *testing.T) {
	assert := assert.New(t)

	serviceName := "my-token"
	config := newConfig(
		WithServiceName(serviceName),
	)

	assert.EqualValues(serviceName, config.options.Tags[ls.ComponentNameKey])
}

func TestWithPlainText(t *testing.T) {
	assert := assert.New(t)

	tests := []bool{
		true,
		false,
	}

	for _, test := range tests {
		config := newConfig(
			WithPlainText(test),
		)
		assert.EqualValues(test, config.options.Collector.Plaintext)
		assert.EqualValues(test, config.options.SystemMetrics.Endpoint.Plaintext)
	}
}

func TestSystemMetricsDisabled(t *testing.T) {
	assert := assert.New(t)

	tests := []bool{
		true,
		false,
	}

	for _, test := range tests {
		config := newConfig(
			WithSystemMetricsDisabled(test),
		)
		assert.EqualValues(test, config.options.SystemMetrics.Disabled)
	}
}

func TestWithSystemMetricTimeout(t *testing.T) {
	assert := assert.New(t)

	tests := []time.Duration{
		1 * time.Second,
		2 * time.Minute,
		3 * time.Hour,
	}

	for _, timeout := range tests {
		config := newConfig(
			WithSystemMetricTimeout(timeout),
		)

		assert.EqualValues(timeout, config.options.SystemMetrics.Timeout)
	}
}

func TestWithSystemMetricMeasurementFrequency(t *testing.T) {
	assert := assert.New(t)

	tests := []time.Duration{
		1 * time.Second,
		2 * time.Minute,
		3 * time.Hour,
	}

	for _, timeout := range tests {
		config := newConfig(
			WithSystemMetricMeasurementFrequency(timeout),
		)

		assert.EqualValues(timeout, config.options.SystemMetrics.MeasurementFrequency)
	}
}
