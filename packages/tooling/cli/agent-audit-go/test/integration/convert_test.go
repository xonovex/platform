//go:build integration

// Package integration holds the tests that read the real repository. They carry
// the integration build tag and run in the ci-integration task, never in
// ci-check.
package integration

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/legacy"
)

// repositoryRoot returns the root of the repository that holds this module.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = <root>/packages/tooling/cli/agent-audit-go/test/integration/convert_test.go
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file)))))))
}

var (
	label  = regexp.MustCompile(`\*\*[^*:]+:\*\*`)
	marker = regexp.MustCompile(`^\s*(#+|[-*])\s+`)
)

// words returns the words of a legacy file that carry content: markdown
// markers, field labels and list punctuation are left out.
func words(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = label.ReplaceAllString(marker.ReplaceAllString(line, ""), " ")
		for _, word := range strings.Fields(line) {
			word = strings.Trim(word, "`,")
			if word != "" {
				out = append(out, word)
			}
		}
	}
	return out
}

// strs collects every string of a decoded JSON value.
func strs(value any, out *[]string) {
	switch v := value.(type) {
	case string:
		*out = append(*out, v)
	case []any:
		for _, item := range v {
			strs(item, out)
		}
	case map[string]any:
		for _, item := range v {
			strs(item, out)
		}
	}
}

func TestEverySourcesFileOfThePluginsConvertsLosslessly(t *testing.T) {
	fsys := osfs.New()
	root := filepath.ToSlash(filepath.Join(repositoryRoot(t), "packages/runtime/plugin"))
	files, err := discovery.Walk(fsys, root, discovery.Options{Skip: discovery.DefaultSkip(), Match: discovery.Named("SOURCES.md")})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 80 {
		t.Fatalf("found %d SOURCES.md files; the catalog has more than 80", len(files))
	}
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		// The templates under assets/ carry placeholders such as {YYYY-MM-DD}
		// that are not valid data; a skill copies one and fills it in.
		if strings.Contains(file, "/assets/") {
			continue
		}
		t.Run(strings.TrimPrefix(file, root+"/"), func(t *testing.T) {
			raw, err := fsys.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			result := legacy.Convert(string(raw), file)
			if len(result.Findings) > 0 {
				t.Fatalf("conversion findings: %v", result.Findings)
			}
			data, err := sources.Encode(result.Document)
			if err != nil {
				t.Fatal(err)
			}
			violations, err := validator.Validate(schemas.Sources, data)
			if err != nil || len(violations) > 0 {
				t.Fatalf("violations = %v, err = %v", violations, err)
			}
			if _, err := sources.Decode(data); err != nil {
				t.Fatal(err)
			}

			var decoded any
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			var kept []string
			strs(decoded, &kept)
			haystack := strings.ReplaceAll(strings.Join(kept, "\n"), "`", "")
			for _, word := range words(string(raw)) {
				if !strings.Contains(haystack, strings.ReplaceAll(word, "`", "")) {
					t.Errorf("the word %q of the legacy file is not in the catalog", word)
				}
			}
		})
	}
}
