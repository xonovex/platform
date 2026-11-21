package config_test

import "embed"

// manifests holds the shipped manifests these tests check in the test binary,
// so a run reads no file and still fails when a manifest drifts from the API.
//
//go:embed samples/*.yaml manager/manager.yaml crd/bases/*.yaml
var manifests embed.FS
