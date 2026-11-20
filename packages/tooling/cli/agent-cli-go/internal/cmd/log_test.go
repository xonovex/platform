package cmd

import (
	"io"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox/plugins"
)

var (
	testLog      = logging.New(io.Discard, io.Discard, false, false)
	testRegistry = plugins.DefaultRegistry(testLog)
)
