// Package wiring is the composition root of the library. It is the one place
// that binds the schema adapter to the embedded schemas and hands the result to
// the packages that need a validator. A command calls it once at start-up.
package wiring

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema/tekuri"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// NewValidator returns a validator that holds every embedded contract schema.
func NewValidator() (jsonschema.Validator, error) {
	return validatorFrom(schemas.Documents)
}

// validatorFrom compiles the documents the source yields. The seam lets a test
// feed a broken schema set.
func validatorFrom(source func() (map[string][]byte, error)) (jsonschema.Validator, error) {
	documents, err := source()
	if err != nil {
		return nil, fmt.Errorf("load contract schemas: %w", err)
	}
	validator, err := tekuri.New(documents)
	if err != nil {
		return nil, fmt.Errorf("compile contract schemas: %w", err)
	}
	return validator, nil
}

// NewSourcesLoader returns a source catalog loader that reads through fsys.
func NewSourcesLoader(fsys fs.FileSystem) (sources.Loader, error) {
	validator, err := NewValidator()
	if err != nil {
		return sources.Loader{}, err
	}
	return sources.Loader{FS: fsys, Validator: validator}, nil
}
