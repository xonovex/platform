package envutil

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseEnv(t *testing.T) {
	got := ParseEnv([]string{"A=1", "B=x=y", "NOEQUALS", "C="})
	want := map[string]string{"A": "1", "B": "x=y", "C": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseEnv = %v, want %v", got, want)
	}
}

func TestParseCustomEnv(t *testing.T) {
	got, err := ParseCustomEnv([]string{"A=1", "B=x=y", "C=", "A=2"})
	if err != nil {
		t.Fatalf("ParseCustomEnv() error = %v", err)
	}
	want := map[string]string{"A": "2", "B": "x=y", "C": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCustomEnv = %v, want %v", got, want)
	}
}

func TestParseCustomEnvRejectsMalformedEntries(t *testing.T) {
	tests := []string{"NOEQUALS", "=value", "1KEY=value", "BAD-NAME=value"}
	for _, entry := range tests {
		t.Run(entry, func(t *testing.T) {
			if _, err := ParseCustomEnv([]string{entry}); err == nil {
				t.Fatalf("ParseCustomEnv(%q) error = nil, want validation error", entry)
			}
		})
	}
}

func TestMergeEnvMaps(t *testing.T) {
	got := MergeEnvMaps(
		map[string]string{"A": "1", "B": "2"},
		map[string]string{"B": "3", "C": "4"},
	)
	want := map[string]string{"A": "1", "B": "3", "C": "4"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MergeEnvMaps = %v, want %v", got, want)
	}
}

func TestEnvMapToSlice(t *testing.T) {
	got := EnvMapToSlice(map[string]string{"A": "1", "B": "2"})
	sort.Strings(got)
	want := []string{"A=1", "B=2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EnvMapToSlice = %v, want %v", got, want)
	}
}
