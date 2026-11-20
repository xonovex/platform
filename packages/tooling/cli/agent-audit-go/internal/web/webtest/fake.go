// Package webtest holds the web test double that the unit tests of this module
// use in place of the network.
package webtest

import (
	"context"
	"fmt"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
)

// Page is the scripted answer for one URL. A page lists the outcomes of its
// successive requests, and the last one repeats.
type Page []Outcome

// Outcome is one answer: a response or an error.
type Outcome struct {
	Response web.Response
	Err      error
}

// Fake answers requests from a script and counts the requests for each URL.
type Fake struct {
	pages map[string]Page
	// Requests counts the requests made for each URL.
	Requests map[string]int
}

// New returns a Fake with the given pages.
func New(pages map[string]Page) *Fake { return &Fake{pages: pages, Requests: map[string]int{}} }

// Get returns the next scripted outcome of the URL. An unscripted URL is an
// error, so a test never passes by accident.
func (f *Fake) Get(_ context.Context, url string) (web.Response, error) {
	page, ok := f.pages[url]
	if !ok || len(page) == 0 {
		return web.Response{}, fmt.Errorf("webtest: no page scripted for %q", url)
	}
	index := f.Requests[url]
	f.Requests[url]++
	if index >= len(page) {
		index = len(page) - 1
	}
	return page[index].Response, page[index].Err
}
