package postgres

import (
	"strings"
	"testing"
)

func TestMigrationContainsDurabilityBoundaries(t *testing.T) {
	content, err := migrations.ReadFile("migrations/001_dispatcher.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, expected := range []string{
		"dispatcher_deliveries",
		"dispatcher_workflow_events",
		"dispatcher_effects",
		"dispatcher_resource_leases",
		"dispatcher_context_records",
		"dispatcher_dead_letters",
		"dispatcher_audit_events",
		"UNIQUE (provider, tenant, idempotency_key)",
		"PRIMARY KEY (provider, tenant, delivery_id)",
	} {
		if !strings.Contains(sql, expected) {
			t.Errorf("migration missing %q", expected)
		}
	}
}
