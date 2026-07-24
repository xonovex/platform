package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestWritePrometheus(t *testing.T) {
	registry := &Registry{}
	registry.DeliveriesAccepted.Add(2)
	registry.EffectsApplied.Add(3)
	var output bytes.Buffer
	if err := registry.WritePrometheus(&output); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"xonovex_dispatcher_deliveries_accepted_total 2",
		"xonovex_dispatcher_effects_applied_total 3",
		"# TYPE xonovex_dispatcher_effects_retried_total counter",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("metrics output missing %q", expected)
		}
	}
}
