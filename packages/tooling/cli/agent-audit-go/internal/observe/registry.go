package observe

import "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"

// Registry holds the transports an audit may use. It names no concrete
// transport: the composition root registers them.
type Registry struct {
	transports []Transport
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds a transport. Transports are consulted in the order they were
// registered.
func (r *Registry) Register(t Transport) *Registry {
	r.transports = append(r.transports, t)
	return r
}

// For returns the transports that can read the source, in registration order.
func (r *Registry) For(source sources.Source) []Transport {
	var out []Transport
	for _, t := range r.transports {
		if t.Applies(source) {
			out = append(out, t)
		}
	}
	return out
}
