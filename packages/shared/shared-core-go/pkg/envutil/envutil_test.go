package envutil

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseCustomEnv(t *testing.T) {
	got := ParseCustomEnv([]string{"A=1", "B=x=y", "NOEQUALS", "C="})
	want := map[string]string{"A": "1", "B": "x=y", "C": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCustomEnv = %v, want %v", got, want)
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
