package lightstep

import (
	"testing"
	"time"

	ls "github.com/lightstep/lightstep-tracer-go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

func TestExport(t *testing.T) {
	assert := assert.New(t)
	now := time.Now().Round(time.Microsecond)
	traceID, _ := core.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := core.SpanIDFromHex("0102030405060708")

	expectedTraceID := uint64(0x102030405060708)
	expectedSpanID := uint64(0x102030405060708)

	tests := []struct {
		name string
		data *trace.SpanData
		want *ls.RawSpan
	}{
		{
			name: "root span",
			data: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: traceID,
					SpanID:  spanID,
				},
				Name:      "/test",
				StartTime: now,
				EndTime:   now,
			},
			want: &ls.RawSpan{
				Context: ls.SpanContext{
					TraceID: expectedTraceID,
					SpanID:  expectedSpanID,
				},
				Operation: "/test",
				Start:     now,
				Duration:  0,
			},
		},
	}

	for _, test := range tests {
		lsSpan := lightStepSpan(test.data)
		assert.EqualValues(test.want.Operation, lsSpan.Operation)
		assert.EqualValues(test.want.Context.SpanID, lsSpan.Context.SpanID)
		assert.EqualValues(test.want.Context.TraceID, lsSpan.Context.TraceID)
		assert.EqualValues(0, lsSpan.ParentSpanID)
	}
}
