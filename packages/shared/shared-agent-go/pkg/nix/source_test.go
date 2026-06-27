package nix

import "testing"

func TestValidatePackageName(t *testing.T) {
	valid := []string{"git", "nodejs_24", "python312Packages.pip", "gcc-wrapper", "fd"}
	for _, name := range valid {
		if !ValidatePackageName(name) {
			t.Errorf("ValidatePackageName(%q) = false, want true", name)
		}
	}
	invalid := []string{"", "with spaces", "semi;colon", "back`tick", "$(inject)"}
	for _, name := range invalid {
		if ValidatePackageName(name) {
			t.Errorf("ValidatePackageName(%q) = true, want false", name)
		}
	}
}

func TestValidateSource_Packages(t *testing.T) {
	if err := ValidateSource(NixSource{Kind: NixSourcePackages, Packages: []string{"git", "ripgrep"}}); err != nil {
		t.Errorf("ValidateSource(packages) = %v, want nil", err)
	}
	// Named sets expand and validate.
	if err := ValidateSource(NixSource{Kind: NixSourcePackages, Packages: []string{"python"}}); err != nil {
		t.Errorf("ValidateSource(set) = %v, want nil", err)
	}
	if err := ValidateSource(NixSource{Kind: NixSourcePackages}); err == nil {
		t.Error("ValidateSource(no packages) = nil, want error")
	}
	if err := ValidateSource(NixSource{Kind: NixSourcePackages, Packages: []string{"bad name"}}); err == nil {
		t.Error("ValidateSource(bad package name) = nil, want error")
	}
}

func TestValidateSource_ProjectFlake(t *testing.T) {
	if err := ValidateSource(NixSource{Kind: NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"}); err != nil {
		t.Errorf("ValidateSource(flake) = %v, want nil", err)
	}
	if err := ValidateSource(NixSource{Kind: NixSourceProjectFlake}); err == nil {
		t.Error("ValidateSource(flake without ref) = nil, want error")
	}
}

func TestValidateSource_UnknownKind(t *testing.T) {
	if err := ValidateSource(NixSource{Kind: "bogus"}); err == nil {
		t.Error("ValidateSource(unknown kind) = nil, want error")
	}
}

func TestComputeEnvID_Stable(t *testing.T) {
	s := NixSource{Kind: NixSourcePackages, Rev: "abc123", Packages: []string{"git", "ripgrep"}}
	if got, want := ComputeEnvID(s), ComputeEnvID(s); got != want {
		t.Errorf("ComputeEnvID not stable: %q != %q", got, want)
	}
	if id := ComputeEnvID(s); len(id) != 16 {
		t.Errorf("ComputeEnvID length = %d, want 16", len(id))
	}
}

func TestComputeEnvID_OrderIndependent(t *testing.T) {
	a := ComputeEnvID(NixSource{Kind: NixSourcePackages, Rev: "abc123", Packages: []string{"git", "ripgrep", "fd"}})
	b := ComputeEnvID(NixSource{Kind: NixSourcePackages, Rev: "abc123", Packages: []string{"fd", "git", "ripgrep"}})
	if a != b {
		t.Errorf("ComputeEnvID is order-dependent: %q != %q", a, b)
	}
}

func TestComputeEnvID_DistinguishesInputs(t *testing.T) {
	base := NixSource{Kind: NixSourcePackages, Rev: "abc123", Packages: []string{"git"}}
	cases := []NixSource{
		{Kind: NixSourcePackages, Rev: "def456", Packages: []string{"git"}},       // rev differs
		{Kind: NixSourcePackages, Rev: "abc123", Packages: []string{"git", "fd"}}, // packages differ
		{Kind: NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"},        // kind differs
	}
	baseID := ComputeEnvID(base)
	for _, c := range cases {
		if ComputeEnvID(c) == baseID {
			t.Errorf("ComputeEnvID collision between %+v and %+v", base, c)
		}
	}
}

func TestComputeEnvID_FlakeShellMatters(t *testing.T) {
	a := ComputeEnvID(NixSource{Kind: NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"})
	b := ComputeEnvID(NixSource{Kind: NixSourceProjectFlake, FlakeRef: "/repo", Shell: "go"})
	if a == b {
		t.Error("ComputeEnvID should distinguish devShell names")
	}
}
