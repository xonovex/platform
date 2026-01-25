// Package scriptlib provides shared utilities for Xonovex tools.
// This package re-exports commonly used functions from sub-packages
// for convenient single-import usage.
//
// For direct package imports, use:
//   - github.com/xonovex/platform/packages/lib/core-go/pkg/colors
//   - github.com/xonovex/platform/packages/lib/core-go/pkg/logging
package scriptlib

import (
	"github.com/xonovex/platform/packages/lib/core-go/pkg/colors"
	"github.com/xonovex/platform/packages/lib/core-go/pkg/logging"
)

// Color functions
var (
	WithRed    = colors.WithRed
	WithGreen  = colors.WithGreen
	WithYellow = colors.WithYellow
	WithBlue   = colors.WithBlue
	WithPurple = colors.WithPurple
	WithCyan   = colors.WithCyan
	WithGray   = colors.WithGray
	WithBold   = colors.WithBold
	Colorize   = colors.Colorize
)

// Color constants
const (
	ColorReset  = colors.Reset
	ColorRed    = colors.Red
	ColorGreen  = colors.Green
	ColorYellow = colors.Yellow
	ColorBlue   = colors.Blue
	ColorPurple = colors.Purple
	ColorCyan   = colors.Cyan
	ColorGray   = colors.Gray
	ColorBold   = colors.Bold
	ColorNC     = colors.NC
)

// Logging functions
var (
	LogInfo         = logging.LogInfo
	LogSuccess      = logging.LogSuccess
	LogWarning      = logging.LogWarning
	LogError        = logging.LogError
	LogDebug        = logging.LogDebug
	PrintSection    = logging.PrintSection
	PrintSubsection = logging.PrintSubsection
	CheckResult     = logging.CheckResult
)

// CheckStatus type and constants
type CheckStatus = logging.CheckStatus

const (
	StatusPass    = logging.StatusPass
	StatusOK      = logging.StatusOK
	StatusSuccess = logging.StatusSuccess
	StatusFail    = logging.StatusFail
	StatusError   = logging.StatusError
	StatusFailed  = logging.StatusFailed
	StatusWarn    = logging.StatusWarn
	StatusWarning = logging.StatusWarning
	StatusInfo    = logging.StatusInfo
)

// IsColorSupported returns true if terminal supports color output
var IsColorSupported = colors.IsColorSupported
