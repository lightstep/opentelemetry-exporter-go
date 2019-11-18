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

	expectedTraceID := uint64(1731642887311460360)
	expectedSpanID := uint64(578437695752307201)

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
		assert.EqualValues(lsSpan.Operation, test.want.Operation)
		assert.EqualValues(lsSpan.Context.SpanID, test.want.Context.SpanID)
		assert.EqualValues(lsSpan.Context.TraceID, test.want.Context.TraceID)
		assert.EqualValues(lsSpan.ParentSpanID, 0)
	}
}
