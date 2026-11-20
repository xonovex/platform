package legacy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// material is the provider content a project of the previous release
// supplies: its provider files and the model overrides of its file. Its
// digest is the digest the previous ledger records, so an approval given
// before this release still holds.
type material struct {
	dir       string
	documents []providers.Document
	models    []string
}

func (m material) empty() bool { return len(m.documents) == 0 && len(m.models) == 0 }

// digest is the tree digest of the provider files, each as
// providers/<name>.json, and of the model overrides as one models file.
func (m material) digest() (string, error) {
	files := make([]artifact.File, 0, len(m.documents)+1)
	for _, document := range m.documents {
		files = append(files, artifact.File{Path: "providers/" + document.Preset.Name + providers.Extension, Content: document.Raw})
	}
	if len(m.models) > 0 {
		files = append(files, artifact.File{Path: "models", Content: []byte(strings.Join(m.models, "\n") + "\n")})
	}
	tree, err := artifact.NewTree(files)
	if err != nil {
		return "", fmt.Errorf("the legacy project providers in %s: %w", m.dir, err)
	}
	return tree.Digest(), nil
}

// review is what a person reads before approving the material: the hosts and
// model slots of each provider file, and the model overrides.
func (m material) review() string {
	var lines []string
	for _, document := range m.documents {
		preset := document.Preset
		slots := make([]string, 0, len(preset.Models))
		for _, name := range slices.Sorted(maps.Keys(preset.Models)) {
			slots = append(slots, name+"="+preset.Models[name].ID)
		}
		lines = append(lines,
			fmt.Sprintf("Provider %s for %s in %s", preset.Name, preset.Agent, document.Origin),
			"  Route hosts: "+orNone(preset.Hosts()),
			"  Model slots: "+orNone(slots),
		)
	}
	if len(m.models) > 0 {
		lines = append(lines, "Model overrides: "+strings.Join(m.models, ", "))
	}
	return strings.Join(lines, "\n") + "\n"
}

func orNone(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

// pending returns the material behind the trust gate, or nothing when the
// previous ledger approves its digest. The new ledger is asked by the gate.
func (m material) pending(host runconfig.Host, origins []string) ([]runconfig.Pending, error) {
	if m.empty() {
		return nil, nil
	}
	digest, err := m.digest()
	if err != nil {
		return nil, err
	}
	approved, err := previouslyApproved(host, digest)
	if err != nil {
		return nil, err
	}
	if approved {
		return nil, nil
	}
	subject := trust.Subject{
		Digest: digest,
		Path:   m.dir + runconfig.LegacySuffix,
		Notice: fmt.Sprintf("The legacy project providers and model overrides in %s take effect only after you approve their digest", m.dir),
		Review: m.review(),
	}
	return []runconfig.Pending{{Subject: subject, Origins: origins}}, nil
}

// ledgerKind is the kind of a project provider approval in the previous
// ledger.
const ledgerKind = "project-providers"

// previousLedger is the approval ledger of the previous release.
type previousLedger struct {
	Version  int `json:"version"`
	Approved []struct {
		Kind    string `json:"kind"`
		Subject string `json:"subject"`
		Digest  string `json:"digest"`
	} `json:"approved"`
}

// ledgerFile is the ledger of the previous release in the data directory.
func ledgerFile(host runconfig.Host) (string, error) {
	data := host.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, err := host.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find the home directory: %w", err)
		}
		data = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(data, "agent-cli", "approvals.json"), nil
}

// previouslyApproved reports whether the ledger of the previous release
// approves the digest. A ledger that does not decode is an error, so a
// damaged ledger approves nothing.
func previouslyApproved(host runconfig.Host, digest string) (bool, error) {
	file, err := ledgerFile(host)
	if err != nil {
		return false, err
	}
	data, err := host.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read the legacy approval ledger %s: %w", file, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var ledger previousLedger
	if err := decoder.Decode(&ledger); err != nil || ledger.Version != 1 {
		return false, fmt.Errorf("the legacy approval ledger %s is not a version 1 ledger", file)
	}
	for _, entry := range ledger.Approved {
		if entry.Kind == ledgerKind && entry.Digest == digest {
			return true, nil
		}
	}
	return false, nil
}
