// Package http is the transport that reads the documents of a source over the
// network, through the web port.
package http

import (
	"context"
	"fmt"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
)

// Name is the name of the transport in reports.
const Name = "http"

// Timeout bounds one request.
const Timeout = 15 * time.Second

// Statuses that mean the provider declined to answer: the address may be fine
// but the audit cannot see it.
const (
	statusUnauthorized = 401
	statusForbidden    = 403
	statusNotFound     = 404
	statusRateLimited  = 429
	statusErrorFloor   = 400
)

// Transport fetches every URL of a source.
type Transport struct {
	Client web.Client
}

// Name returns the name of the transport.
func (Transport) Name() string { return Name }

// Remote returns true: reading a URL uses the network.
func (Transport) Remote() bool { return true }

// Applies reports whether the source lists a URL.
func (Transport) Applies(source sources.Source) bool { return len(source.URLs) > 0 }

// Observe fetches the URLs in order. A request that fails is tried once more
// before it counts.
func (t Transport) Observe(ctx context.Context, request observe.Request) (observe.Observation, error) {
	documents := make([]observe.Document, 0, len(request.Source.URLs))
	for _, url := range request.Source.URLs {
		document := t.fetch(ctx, url)
		if document.Status == observe.Failed {
			document = t.fetch(ctx, url)
		}
		documents = append(documents, document)
	}
	return observe.Observation{Transport: Name, Documents: documents}, nil
}

func (t Transport) fetch(ctx context.Context, url string) observe.Document {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	document := observe.Document{Locator: url}
	response, err := t.Client.Get(ctx, url)
	if err != nil {
		document.Status = observe.Failed
		document.Detail = "unreachable: " + err.Error()
		return document
	}
	document.Detail = fmt.Sprintf("HTTP %d", response.Status)
	if response.LastModified != "" {
		document.Detail += ", Last-Modified: " + response.LastModified
	}
	switch {
	case response.Status == statusNotFound:
		document.Status = observe.Missing
		document.Detail = fmt.Sprintf("HTTP %d", response.Status)
	case response.Status == statusUnauthorized, response.Status == statusForbidden, response.Status == statusRateLimited:
		document.Status = observe.Restricted
	case response.Status <= 0 || response.Status >= statusErrorFloor:
		document.Status = observe.Failed
	default:
		document.Status = observe.OK
		document.SHA256 = observe.Hash(response.Body)
		document.Body = response.Body
	}
	return document
}
