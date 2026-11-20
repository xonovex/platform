// Package nethttp is the web adapter that reads documents over the network.
// Only the composition root imports it.
package nethttp

import (
	"context"
	"io"
	"net/http"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
)

// userAgent names this tool to the servers it reads.
const userAgent = "agent-audit/1"

// maxBody bounds the bytes read from one document.
const maxBody = 32 << 20

// Client reads URLs with an http.Client.
type Client struct {
	http *http.Client
}

// New returns a Client that uses the default transport.
func New() Client { return Client{http: &http.Client{}} }

// Get requests the URL. The body is read only for a status below 400.
func (c Client) Get(ctx context.Context, url string) (web.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return web.Response{}, err
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := c.http.Do(request)
	if err != nil {
		return web.Response{}, err
	}
	defer func() { _ = response.Body.Close() }()
	out := web.Response{Status: response.StatusCode, LastModified: response.Header.Get("Last-Modified")}
	if response.StatusCode < 400 {
		body, err := io.ReadAll(io.LimitReader(response.Body, maxBody))
		if err != nil {
			return web.Response{}, err
		}
		out.Body = body
	}
	return out, nil
}
