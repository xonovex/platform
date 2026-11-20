// Package web is the port for reading a document over HTTP. A core package asks
// for a URL through Client, so a unit test answers from webtest.Fake and only
// the nethttp adapter touches the network.
package web

import "context"

// Response is the answer to a request.
type Response struct {
	// Status is the HTTP status code.
	Status int
	// LastModified is the Last-Modified header, empty when the server sent none.
	LastModified string
	// Body is the content. An adapter reads it only for a status below 400, so
	// an error page is never downloaded.
	Body []byte
}

// Client reads one URL. A status of any value is a Response. The error is for a
// request that did not complete, such as a refused connection or a timeout.
type Client interface {
	Get(ctx context.Context, url string) (Response, error)
}
