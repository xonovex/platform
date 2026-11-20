// Package explain renders the merged configuration with the layer and the file
// that gave each value, the output of a config show --sources command.
package explain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Render writes one line per value, and one line per entry of a Merge key:
// the dotted name, the compact JSON value and its origins.
func Render(w io.Writer, resolved merge.Resolved) error {
	for _, value := range resolved.Values() {
		name := value.Key.Name()
		if value.Key.Rule != spec.Merge {
			if _, err := fmt.Fprintf(w, "%s = %s  (%s)\n", name, compact(value.Raw), origins(value.Origins)); err != nil {
				return err
			}
			continue
		}
		var entries map[string]json.RawMessage
		// A merged value is always an object, which the merge encoded.
		_ = json.Unmarshal(value.Raw, &entries)
		names := make([]string, 0, len(entries))
		for entry := range entries {
			names = append(names, entry)
		}
		sort.Strings(names)
		for _, entry := range names {
			if _, err := fmt.Fprintf(w, "%s.%s = %s  (%s)\n", name, entry, compact(entries[entry]), origins([]merge.Origin{value.Entries[entry]})); err != nil {
				return err
			}
		}
	}
	return nil
}

// compact drops the white space of a merged value, which is always valid JSON.
func compact(raw json.RawMessage) string {
	var out bytes.Buffer
	_ = json.Compact(&out, raw)
	return out.String()
}

func origins(list []merge.Origin) string {
	parts := make([]string, len(list))
	for i, origin := range list {
		parts[i] = string(origin.Layer)
		if origin.Path != "" {
			parts[i] += " " + origin.Path
		}
	}
	if len(parts) == 0 {
		return "unset"
	}
	return strings.Join(parts, ", ")
}
