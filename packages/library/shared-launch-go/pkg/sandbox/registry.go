// Package sandbox is the selection layer of the confinement axes: the Registry
// of lazy plugin factories, Evaluate, which computes the guarantees of a
// request, and Select, which enforces a policy on them fail-closed. It names no
// concrete isolation or provisioner; a composition root registers them.
package sandbox

import (
	"errors"
	"fmt"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

var (
	// ErrNoIsolator reports an isolation method with no registered isolator.
	ErrNoIsolator = errors.New("no isolator registered")
	// ErrNoProvisioner reports a provisioning method with no registered provisioner.
	ErrNoProvisioner = errors.New("no provisioner registered")
)

// IsolatorFactory and ProvisionerFactory construct a fresh plugin instance.
type (
	IsolatorFactory[I isoshared.Confinement] func() I
	ProvisionerFactory                       func() provshared.Provisioner
)

// Registry maps method names to plugin factories. It is built once at the
// composition root and passed explicitly to Select, so there is no global mutable
// state, and tests build their own minimal Registry. Adding a plugin is one
// Register call at the root; the core selection/policy code never changes.
//
// I is the isolator type of the executor: selection needs only its
// Confinement, and the registry hands the executor back the full I it
// registered, so an executor that launches nothing itself registers
// Confinement alone.
type Registry[I isoshared.Confinement] struct {
	isolators    map[isolation.IsolationMethod]IsolatorFactory[I]
	provisioners map[provision.ProvisionMethod]ProvisionerFactory
}

// NewRegistry returns an empty registry.
func NewRegistry[I isoshared.Confinement]() *Registry[I] {
	return &Registry[I]{
		isolators:    map[isolation.IsolationMethod]IsolatorFactory[I]{},
		provisioners: map[provision.ProvisionMethod]ProvisionerFactory{},
	}
}

// RegisterIsolator registers (or replaces) an isolator factory; chainable.
func (r *Registry[I]) RegisterIsolator(m isolation.IsolationMethod, f IsolatorFactory[I]) *Registry[I] {
	r.isolators[m] = f
	return r
}

// RegisterProvisioner registers (or replaces) a provisioner factory; chainable.
func (r *Registry[I]) RegisterProvisioner(m provision.ProvisionMethod, f ProvisionerFactory) *Registry[I] {
	r.provisioners[m] = f
	return r
}

// Isolator constructs the isolator for m, or ErrNoIsolator if unregistered.
func (r *Registry[I]) Isolator(m isolation.IsolationMethod) (I, error) {
	f, ok := r.isolators[m]
	if !ok {
		var zero I
		return zero, fmt.Errorf("%w for isolation %q", ErrNoIsolator, m)
	}
	return f(), nil
}

// Provisioner constructs the provisioner for m, or ErrNoProvisioner if unregistered.
func (r *Registry[I]) Provisioner(m provision.ProvisionMethod) (provshared.Provisioner, error) {
	f, ok := r.provisioners[m]
	if !ok {
		return nil, fmt.Errorf("%w for provisioning %q", ErrNoProvisioner, m)
	}
	return f(), nil
}

// IsolationMethods returns the registered isolation methods, sorted for stable output.
func (r *Registry[I]) IsolationMethods() []isolation.IsolationMethod {
	out := make([]isolation.IsolationMethod, 0, len(r.isolators))
	for m := range r.isolators {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Request bundles the per-run axis selection handed to Select.
type Request struct {
	Isolation      isolation.IsolationMethod
	Provision      provision.ProvisionMethod
	Network        network.NetworkMethod
	Passthrough    bool
	Runtime        string
	Image          string
	HasCustomBinds bool
}

// Choice is the sandbox a Request resolves to: the isolator and the provisioner
// with the guarantees they establish together for that request.
type Choice[I isoshared.Confinement] struct {
	Isolator     I
	Provisioner  provshared.Provisioner
	Capabilities policy.Capabilities
}

// Evaluate resolves the isolator and provisioner for a request and computes the
// guarantees they establish, without demanding any. Registry membership is the
// validity check; the resolved plugins declare their own capabilities, so this
// layer never names a concrete isolator or provisioner. Bind paths fail for an
// isolator that cannot bind a host path, a network mode that needs confinement
// fails for an isolator that cannot confine the network, and a managed run fails
// for an isolator that declares no egress transport, whatever the policy.
func Evaluate[I isoshared.Confinement](reg *Registry[I], req Request) (Choice[I], error) {
	iso, err := reg.Isolator(req.Isolation)
	if err != nil {
		return Choice[I]{}, err
	}
	prov, err := reg.Provisioner(req.Provision)
	if err != nil {
		return Choice[I]{}, err
	}
	if req.HasCustomBinds && !iso.AcceptsHostBinds() {
		return Choice[I]{}, fmt.Errorf("isolation=%s cannot bind a host path; remove the bind paths", req.Isolation)
	}
	if network.NeedsConfinement(req.Network) && !iso.ConfinesNetwork() {
		return Choice[I]{}, fmt.Errorf("isolation=%s cannot restrict network egress (network=%q); select an isolation that confines the network", req.Isolation, req.Network)
	}
	transport := iso.EgressTransport()
	if err := network.CheckManaged(req.Network, transport); err != nil {
		return Choice[I]{}, fmt.Errorf("isolation=%s: %w", req.Isolation, err)
	}
	return Choice[I]{
		Isolator:    iso,
		Provisioner: prov,
		Capabilities: policy.Capabilities{
			Pinned:               iso.PinnedProvision(req.Provision, prov.Pinned(), req.Image),
			HostToolsUnreachable: iso.HidesHost(req.Passthrough, req.Image) && !req.HasCustomBinds,
			EgressRestricted:     network.EgressRestricted(req.Network, transport),
			KernelIsolated:       iso.KernelIsolated(req.Runtime),
		},
	}, nil
}

// Select resolves the choice of a request and enforces the policy fail-closed:
// it refuses a request whose guarantees fall short of what the policy demands.
func Select[I isoshared.Confinement](reg *Registry[I], req Request, pol policy.SandboxPolicy) (Choice[I], error) {
	choice, err := Evaluate(reg, req)
	if err != nil {
		return Choice[I]{}, err
	}
	if err := policy.EnforcePolicy(choice.Capabilities, pol); err != nil {
		return Choice[I]{}, err
	}
	return choice, nil
}

// AvailableIsolations returns the registered isolation methods whose isolator is
// currently available on this host. It needs the Launcher of each isolator,
// because availability is a probe of the machine that launches it.
func AvailableIsolations[I isoshared.Isolator](reg *Registry[I]) []isolation.IsolationMethod {
	var available []isolation.IsolationMethod
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
