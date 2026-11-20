package report_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/report"
)

func TestParseAcceptsOnlyTheSupportedFormats(t *testing.T) {
	got, err := report.Parse("sarif", report.Text, report.SARIF)
	if err != nil || got != report.SARIF {
		t.Fatalf("Parse = %q, %v", got, err)
	}
	_, err = report.Parse("json", report.Text, report.SARIF)
	if !errors.Is(err, report.ErrUnknownFormat) || err.Error() != `unknown output format "json": use text or sarif` {
		t.Fatalf("err = %v", err)
	}
}

func TestNames(t *testing.T) {
	cases := map[string][]report.Format{
		"":                    nil,
		"text":                {report.Text},
		"text or json":        {report.Text, report.JSON},
		"text, json or sarif": {report.Text, report.JSON, report.SARIF},
	}
	for want, formats := range cases {
		if got := report.Names(formats...); got != want {
			t.Errorf("Names(%v) = %q, want %q", formats, got, want)
		}
	}
}
