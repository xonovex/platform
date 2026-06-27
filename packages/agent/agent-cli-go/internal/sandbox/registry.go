package sandbox

import (
	"errors"
	"fmt"
	"sort"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

var (
	// ErrNoIsolator reports an isolation method with no registered isolator.
	ErrNoIsolator = errors.New("no isolator registered")
	// ErrNoProvisioner reports a provisioning method with no registered provisioner.
	ErrNoProvisioner = errors.New("no provisioner registered")
)

// IsolatorFactory and ProvisionerFactory construct a fresh plugin instance.
type (
	IsolatorFactory    func() Isolator
	ProvisionerFactory func() Provisioner
)

// Registry maps method names to plugin factories. It is built once at the
// composition root and passed explicitly to Select — there is no global mutable
// state, and tests build their own minimal Registry. Adding a plugin is one
// Register call at the root; the core selection/policy code never changes.
type Registry struct {
	isolators    map[types.IsolationMethod]IsolatorFactory
	provisioners map[types.ProvisionMethod]ProvisionerFactory
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		isolators:    map[types.IsolationMethod]IsolatorFactory{},
		provisioners: map[types.ProvisionMethod]ProvisionerFactory{},
	}
}

// RegisterIsolator registers (or replaces) an isolator factory; chainable.
func (r *Registry) RegisterIsolator(m types.IsolationMethod, f IsolatorFactory) *Registry {
	r.isolators[m] = f
	return r
}

// RegisterProvisioner registers (or replaces) a provisioner factory; chainable.
func (r *Registry) RegisterProvisioner(m types.ProvisionMethod, f ProvisionerFactory) *Registry {
	r.provisioners[m] = f
	return r
}

// Isolator constructs the isolator for m, or ErrNoIsolator if unregistered.
func (r *Registry) Isolator(m types.IsolationMethod) (Isolator, error) {
	f, ok := r.isolators[m]
	if !ok {
		return nil, fmt.Errorf("%w for isolation %q", ErrNoIsolator, m)
	}
	return f(), nil
}

// Provisioner constructs the provisioner for m, or ErrNoProvisioner if unregistered.
func (r *Registry) Provisioner(m types.ProvisionMethod) (Provisioner, error) {
	f, ok := r.provisioners[m]
	if !ok {
		return nil, fmt.Errorf("%w for provisioning %q", ErrNoProvisioner, m)
	}
	return f(), nil
}

// IsolationMethods returns the registered isolation methods, sorted for stable output.
func (r *Registry) IsolationMethods() []types.IsolationMethod {
	out := make([]types.IsolationMethod, 0, len(r.isolators))
	for m := range r.isolators {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Request bundles the per-run axis selection handed to Select.
type Request struct {
	Isolation   types.IsolationMethod
	Provision   types.ProvisionMethod
	Network     types.NetworkMethod
	Passthrough bool
	Runtime     string
	Image       string
}

// Select resolves the isolator and provisioner for a request and enforces the
// policy fail-closed. Registry membership is the validity check; the resolved
// plugins declare their own capabilities, so the policy engine never names a
// concrete isolator or provisioner.
func Select(reg *Registry, req Request, pol types.SandboxPolicy) (Isolator, Provisioner, error) {
	iso, err := reg.Isolator(req.Isolation)
	if err != nil {
		return nil, nil, err
	}
	prov, err := reg.Provisioner(req.Provision)
	if err != nil {
		return nil, nil, err
	}
	caps := sandbox.Capabilities{
		Pinned:               prov.Pinned(),
		HostToolsUnreachable: iso.HidesHost(req.Passthrough, req.Image),
		EgressRestricted:     sandbox.EgressIsRestricted(req.Network),
		KernelIsolated:       iso.KernelIsolated(req.Runtime),
	}
	if err := sandbox.EnforcePolicy(caps, pol); err != nil {
		return nil, nil, err
	}
	return iso, prov, nil
}

// AvailableIsolations returns the registered isolation methods whose isolator is
// currently available on this host.
func AvailableIsolations(reg *Registry) []types.IsolationMethod {
	var available []types.IsolationMethod
	for _, m := range reg.IsolationMethods() {
		iso, err := reg.Isolator(m)
		if err != nil {
			continue
		}
		if ok, err := iso.Available(); err == nil && ok {
			available = append(available, m)
		}
	}
	return available
}
