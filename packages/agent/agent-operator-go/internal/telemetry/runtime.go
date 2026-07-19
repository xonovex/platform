package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Signal is deliberately closed and content-minimized. It has no prompt,
// secret, arbitrary-message, or arbitrary-attribute field.
type Signal struct {
	CorrelationID     string
	RunReference      string
	Kind              string
	Control           string
	Outcome           string
	EffectiveAutonomy string
}

// Sink is the telemetry output port used by the runtime.
type Sink interface {
	Record(context.Context, Signal)
}

// OTelSink emits a span and a low-cardinality counter through the globally
// configured OpenTelemetry providers. Exporter choice remains deployment-owned.
type OTelSink struct {
	tracer  trace.Tracer
	counter metric.Int64Counter
}

func NewOTelSink() *OTelSink {
	meter := otel.Meter("github.com/xonovex/agent-operator-go")
	counter, _ := meter.Int64Counter("xonovex.governance.events")
	return &OTelSink{
		tracer:  otel.Tracer("github.com/xonovex/agent-operator-go"),
		counter: counter,
	}
}

func (sink *OTelSink) Record(ctx context.Context, signal Signal) {
	attributes := signalAttributes(signal)
	_, span := sink.tracer.Start(ctx, "xonovex."+signal.Kind, trace.WithAttributes(attributes...))
	if sink.counter != nil {
		sink.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("xonovex.signal.kind", signal.Kind),
			attribute.String("xonovex.signal.outcome", signal.Outcome),
		))
	}
	span.End()
}

func signalAttributes(signal Signal) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("xonovex.correlation.id", signal.CorrelationID),
		attribute.String("xonovex.run.reference", signal.RunReference),
		attribute.String("xonovex.signal.kind", signal.Kind),
		attribute.String("xonovex.oversight.control", signal.Control),
		attribute.String("xonovex.signal.outcome", signal.Outcome),
		attribute.String("xonovex.autonomy.effective", signal.EffectiveAutonomy),
	}
}

// DriftAssessment is the deterministic result consumed by containment.
type DriftAssessment struct {
	Detected          bool
	EffectiveAutonomy string
	FailureCodes      []string
}

// AssessA3 recomputes effective autonomy exclusively from observed control
// signals. Any required control that is missing or degraded demotes A3 to A2.
func AssessA3(signals []Signal, requiredControls []string) DriftAssessment {
	outcomes := make(map[string]string, len(signals))
	for _, signal := range signals {
		if signal.Kind == "oversight.control" {
			outcomes[signal.Control] = signal.Outcome
		}
	}
	assessment := DriftAssessment{EffectiveAutonomy: "A3"}
	for _, control := range requiredControls {
		if outcomes[control] != "healthy" {
			assessment.Detected = true
			assessment.EffectiveAutonomy = "A2"
			assessment.FailureCodes = append(assessment.FailureCodes, fmt.Sprintf("oversight-%s-degraded", control))
		}
	}
	return assessment
}
