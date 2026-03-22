package nix

import (
	"slices"
	"testing"
)

func TestExpandPackageSets_PlainPackages(t *testing.T) {
	result := ExpandPackageSets([]string{"git", "curl"})
	expected := []string{"git", "curl"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(plain) = %v, want %v", result, expected)
	}
}

func TestExpandPackageSets_NamedSet(t *testing.T) {
	result := ExpandPackageSets([]string{"go"})
	expected := []string{"go"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(go set) = %v, want %v", result, expected)
	}

	result = ExpandPackageSets([]string{"python"})
	expected = []string{"python312", "python312Packages.pip"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(python set) = %v, want %v", result, expected)
	}
}

func TestExpandPackageSets_UnknownPreserved(t *testing.T) {
	result := ExpandPackageSets([]string{"my-custom-pkg"})
	expected := []string{"my-custom-pkg"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(unknown) = %v, want %v", result, expected)
	}
}

func TestExpandPackageSets_Deduplication(t *testing.T) {
	result := ExpandPackageSets([]string{"git", "git", "curl"})
	expected := []string{"git", "curl"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(dupes) = %v, want %v", result, expected)
	}
}

func TestExpandPackageSets_MixedSetAndPackage(t *testing.T) {
	result := ExpandPackageSets([]string{"python312", "python"})
	// python312 appears first as plain, then python set expands but python312 is deduped
	expected := []string{"python312", "python312Packages.pip"}
	if !slices.Equal(result, expected) {
		t.Errorf("ExpandPackageSets(mixed) = %v, want %v", result, expected)
	}
}

func TestExpandPackageSets_Empty(t *testing.T) {
	result := ExpandPackageSets([]string{})
	if len(result) != 0 {
		t.Errorf("ExpandPackageSets(empty) = %v, want empty", result)
	}
}

func TestValidatePin_Known(t *testing.T) {
	for _, pin := range []string{"nixos-24.11", "nixos-unstable", "nixpkgs-unstable"} {
		if err := ValidatePin(pin); err != nil {
			t.Errorf("ValidatePin(%q) = %v, want nil", pin, err)
		}
	}
}

func TestValidatePin_Empty(t *testing.T) {
	if err := ValidatePin(""); err != nil {
		t.Errorf("ValidatePin(\"\") = %v, want nil", err)
	}
}

func TestValidatePin_Unknown(t *testing.T) {
	if err := ValidatePin("nixos-99.99"); err == nil {
		t.Error("ValidatePin(unknown) = nil, want error")
	}
}
