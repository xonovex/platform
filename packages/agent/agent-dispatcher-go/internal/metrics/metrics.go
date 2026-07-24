package metrics

import (
	"fmt"
	"io"
	"sync/atomic"
)

type Registry struct {
	DeliveriesAccepted  atomic.Uint64
	DeliveriesDuplicate atomic.Uint64
	DeliveriesRejected  atomic.Uint64
	EffectsQueued       atomic.Uint64
	EffectsPreviewed    atomic.Uint64
	EffectsApplied      atomic.Uint64
	EffectsReconciled   atomic.Uint64
	EffectsRetried      atomic.Uint64
	EffectsDeadLettered atomic.Uint64
}

func (registry *Registry) WritePrometheus(writer io.Writer) error {
	metrics := []struct {
		name  string
		value uint64
	}{
		{"xonovex_dispatcher_deliveries_accepted_total", registry.DeliveriesAccepted.Load()},
		{"xonovex_dispatcher_deliveries_duplicate_total", registry.DeliveriesDuplicate.Load()},
		{"xonovex_dispatcher_deliveries_rejected_total", registry.DeliveriesRejected.Load()},
		{"xonovex_dispatcher_effects_queued_total", registry.EffectsQueued.Load()},
		{"xonovex_dispatcher_effects_previewed_total", registry.EffectsPreviewed.Load()},
		{"xonovex_dispatcher_effects_applied_total", registry.EffectsApplied.Load()},
		{"xonovex_dispatcher_effects_reconciled_total", registry.EffectsReconciled.Load()},
		{"xonovex_dispatcher_effects_retried_total", registry.EffectsRetried.Load()},
		{"xonovex_dispatcher_effects_dead_lettered_total", registry.EffectsDeadLettered.Load()},
	}
	for _, metric := range metrics {
		if _, err := fmt.Fprintf(writer, "# TYPE %s counter\n%s %d\n", metric.name, metric.name, metric.value); err != nil {
			return err
		}
	}
	return nil
}
