package launch

import (
	"errors"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
)

// prepareOverlay builds the overlay the request asks for. A run without one
// gets an empty result whose Cleanup does nothing.
func prepareOverlay(overlays Overlays, request *OverlayRequest) (OverlayResult, error) {
	if request == nil {
		return OverlayResult{Cleanup: func() {}}, nil
	}
	if overlays == nil {
		return OverlayResult{}, errors.New("the overlay is not available in this build")
	}
	return overlays.Prepare(*request)
}

// withOverlay builds the overlay of the request, runs body over it and then
// releases it. The release runs on every exit of body, and the outcome names
// the overlay sources whether or not body fails.
func withOverlay(overlays Overlays, log logging.Logger, request *OverlayRequest, body func(OverlayResult) (launched, error)) (result outcome, err error) {
	built, err := prepareOverlay(overlays, request)
	if err != nil {
		return outcome{}, err
	}
	result.overlay = built.Sources
	defer func() { releaseOverlay(log, built, result.runsOn) }()
	result.launched, err = body(built)
	return result, err
}

// releaseOverlay removes the overlay, unless a program that reads it runs on in
// a terminal session after this process returns: that overlay stays on disk,
// and the user is told where.
func releaseOverlay(log logging.Logger, built OverlayResult, runsOn bool) {
	if runsOn && built.Root != "" {
		log.Info("The overlay stays at " + built.Root + " while the terminal session runs; remove it afterwards")
		return
	}
	built.Cleanup()
}
