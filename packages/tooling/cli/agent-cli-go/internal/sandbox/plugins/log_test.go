package plugins

import (
	"io"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)
