package cmd

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
)

func TestSchemaPrintsEachSchemaAsJSON(t *testing.T) {
	for name, id := range map[string]string{"config": config.SchemaID, "provider": providers.SchemaID, "run-record": record.SchemaID} {
		r := newTestRoot()

		if err := r.execute("schema", name); err != nil {
			t.Fatalf("schema %s error = %v", name, err)
		}

		var document map[string]any
		if err := json.Unmarshal(r.out.Bytes(), &document); err != nil || document["$id"] != id {
			t.Errorf("schema %s = $id %v, %v; want %s", name, document["$id"], err, id)
		}
		if !strings.HasSuffix(r.out.String(), "\n") {
			t.Errorf("schema %s output does not end with a newline", name)
		}
	}
}

func TestSchemaListsTheNamesWithoutAnArgument(t *testing.T) {
	r := newTestRoot()

	if err := r.execute("schema"); err != nil {
		t.Fatalf("schema error = %v", err)
	}
	if r.out.String() != "config\nprovider\nrun-record\n" {
		t.Errorf("schema = %q, want the names", r.out.String())
	}
}

func TestSchemaRejectsAnUnknownNameAsUsage(t *testing.T) {
	r := newTestRoot()

	err := r.execute("schema", "nosuch")

	var usage UsageError
	if !errors.As(err, &usage) || !strings.Contains(err.Error(), "config, provider, run-record") {
		t.Errorf("schema nosuch error = %v, want a usage error that lists the schemas", err)
	}
}

func TestSchemaReportsAFailureToWrite(t *testing.T) {
	for _, args := range [][]string{{}, {"config"}} {
		command := newSchemaCommand(brokenWriter{})
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("schema %v error = nil, want the write failure", args)
		}
	}
}

func TestWriteJSONReportsAnUnencodableValue(t *testing.T) {
	if err := writeJSON(brokenWriter{}, make(chan int)); err == nil {
		t.Error("writeJSON(chan) error = nil, want an encoding failure")
	}
}
