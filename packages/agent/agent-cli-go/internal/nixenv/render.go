package nixenv

import (
	"fmt"
	"strings"
)

// GetNixpkgsTarballURL returns the nixpkgs tarball URL for a pin
func GetNixpkgsTarballURL(pin string) (string, error) {
	pinConfig, ok := NixpkgsPins[pin]
	if !ok {
		return "", fmt.Errorf("unknown nixpkgs pin: %s", pin)
	}
	return fmt.Sprintf("https://github.com/NixOS/nixpkgs/archive/%s.tar.gz", pinConfig.Ref), nil
}

// RenderNixExpression renders a Nix expression for building an agent environment
func RenderNixExpression(spec *EnvSpec, envID string) (string, error) {
	tarballURL, err := GetNixpkgsTarballURL(spec.NixpkgsPin)
	if err != nil {
		return "", err
	}

	// Format packages list with proper indentation
	var packagesLines strings.Builder
	for _, pkg := range spec.Packages {
		packagesLines.WriteString("    ")
		packagesLines.WriteString(pkg)
		packagesLines.WriteString("\n")
	}

	return fmt.Sprintf(`# Auto-generated agent environment - do not edit
# EnvID: %s
# Pin: %s

{ pkgs ? import (fetchTarball "%s") {
    config.allowUnfree = true;
  }
}:

pkgs.buildEnv {
  name = "agent-env-%s";
  paths = with pkgs; [
%s  ];
}
`, envID, spec.NixpkgsPin, tarballURL, envID, packagesLines.String()), nil
}
