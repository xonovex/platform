// Package openshell is the isolation=openshell leaf: the agent runs in an
// OpenShell sandbox on a gateway. The leaf writes a policy document for the run,
// creates the sandbox from the image under that policy, uploads the work
// directory, executes the agent, downloads the work directory back and deletes
// the sandbox.
package openshell

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/hostlaunch"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// program is the OpenShell command line tool.
const program = "openshell"

// namePrefix starts the name of every sandbox this leaf creates.
const namePrefix = "agent-cli-"

// maxNames bounds the search for a free sandbox name.
const maxNames = 1000

// sandboxRoot is the directory in the sandbox that holds the uploaded work
// directory.
const sandboxRoot = "/sandbox"

// ErrUnsupported reports a run that an OpenShell sandbox cannot hold.
var ErrUnsupported = errors.New("openshell cannot run this")

// Host is what the leaf needs of the machine: the isolation host plus a private
// place to write the policy document.
type Host interface {
	hostlaunch.Host
	MkdirTemp(pattern string) (string, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	RemoveAll(path string) error
}

// Isolator runs the agent in an OpenShell sandbox.
//
// The gateway enforces the network policy of the sandbox, so the isolator
// declares the policy transport for managed egress: the allowlist is the policy,
// and no proxy runs on the host. The sandbox is created from an image and sees no
// host tool and no host home directory.
type Isolator struct {
	host Host
	log  logging.Logger
}

// NewIsolator creates an OpenShell isolator over host, writing through log.
func NewIsolator(host Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available reports whether the openshell tool is installed.
func (i *Isolator) Available() (bool, error) {
	_, err := i.host.LookPath(program)
	return err == nil, nil
}

// HidesHost reports true for a digest-pinned image: the sandbox runs the image
// tools and mounts nothing of the host.
func (i *Isolator) HidesHost(_ bool, image string) bool {
	return isolation.IsDigestPinnedImage(image)
}

// PinnedProvision requires a digest-pinned image, and a pinned provisioner when
// one runs.
func (i *Isolator) PinnedProvision(method provision.ProvisionMethod, provisionerPinned bool, image string) bool {
	if !isolation.IsDigestPinnedImage(image) {
		return false
	}
	return method == provision.ProvisionNone || provisionerPinned
}

// KernelIsolated reports false: the isolator does not verify the runtime of the
// gateway.
func (i *Isolator) KernelIsolated(string) bool { return false }

// ConfinesNetwork reports true: the policy denies every host it does not list.
func (i *Isolator) ConfinesNetwork() bool { return true }

// EgressTransport declares the policy: the gateway enforces the allowlist.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportPolicy }

// Bounded reports true: the sandbox on the gateway confines what the agent
// writes and reaches.
func (i *Isolator) Bounded() bool { return true }

// AcceptsHostBinds reports false: a sandbox on a gateway has no host bind, so
// the work directory is uploaded instead.
func (i *Isolator) AcceptsHostBinds() bool { return false }

// CheckAgent requires an image, because the sandbox starts from it and the agent
// must be in it.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if cfg.Image == "" {
		return fmt.Errorf("%w: pass --image with an image that contains agent executable %q", ErrUnsupported, cfg.Agent.Binary)
	}
	return nil
}

// plan holds the commands of one run.
type plan struct {
	name     string
	policy   []byte
	create   func(policyFile string) []string
	upload   []string
	exec     []string
	download []string
	remove   []string
}

// Command returns the exec command of the run, for display.
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	p, err := i.plan(cfg, c, namePrefix+"1")
	if err != nil {
		return nil, err
	}
	return append([]string{program}, p.exec...), nil
}

// TerminalCommand cannot place the sandbox lifecycle inside a terminal wrapper.
func (i *Isolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string, error) {
	return nil, nil, fmt.Errorf("%w: a terminal wrapper cannot create and remove the sandbox", ErrUnsupported)
}

// Run creates the sandbox under the policy, uploads the work directory, runs the
// agent, downloads the work directory and deletes the sandbox. It deletes the
// sandbox and the policy file on every path that leaves Run after the policy file
// exists.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	if err := i.CheckAgent(cfg); err != nil {
		return 1, err
	}
	name, err := i.freeName()
	if err != nil {
		return 1, err
	}
	p, err := i.plan(cfg, c, name)
	if err != nil {
		return 1, err
	}
	dir, err := i.host.MkdirTemp("agent-cli-openshell-*")
	if err != nil {
		return 1, fmt.Errorf("create policy directory: %w", err)
	}
	defer func() { _ = i.host.RemoveAll(dir) }()
	policyFile := filepath.Join(dir, "policy.json")
	if err := i.host.WriteFile(policyFile, p.policy, 0o600); err != nil {
		return 1, fmt.Errorf("write sandbox policy: %w", err)
	}

	if _, err := i.host.Capture(program, p.create(policyFile), ""); err != nil {
		_ = i.remove(p)
		return 1, fmt.Errorf("create sandbox: %w", err)
	}
	code, runErr := i.session(p)
	if err := i.remove(p); err != nil {
		i.log.Warning(fmt.Sprintf("Could not delete sandbox %s: %v", name, err))
	}
	return code, runErr
}

// session uploads the work directory, runs the agent, and downloads the work
// directory back so the changes of the agent survive the sandbox. The download
// runs even when the agent fails, because its partial work is still the result.
func (i *Isolator) session(p plan) (int, error) {
	if _, err := i.host.Capture(program, p.upload, ""); err != nil {
		return 1, fmt.Errorf("upload work directory: %w", err)
	}
	code, runErr := hostlaunch.SpawnSandbox(i.host, i.log, program, p.exec, nil, "", "OpenShell sandbox")
	if _, err := i.host.Capture(program, p.download, ""); err != nil {
		return code, errors.Join(runErr, fmt.Errorf("download work directory: %w", err))
	}
	return code, runErr
}

func (i *Isolator) remove(p plan) error {
	_, err := i.host.Capture(program, p.remove, "")
	return err
}

func (i *Isolator) freeName() (string, error) {
	listing, err := i.host.Capture(program, []string{"sandbox", "list", "--names"}, "")
	if err != nil {
		return "", fmt.Errorf("list sandboxes: %w", err)
	}
	taken := map[string]bool{}
	for _, line := range strings.Fields(listing) {
		taken[line] = true
	}
	for n := 1; n <= maxNames; n++ {
		if name := fmt.Sprintf("%s%d", namePrefix, n); !taken[name] {
			return name, nil
		}
	}
	return "", fmt.Errorf("no free sandbox name under %s", namePrefix)
}

func (i *Isolator) plan(cfg isoshared.RunConfig, c provision.Contribution, name string) (plan, error) {
	switch {
	case cfg.Image == "":
		return plan{}, fmt.Errorf("%w: the sandbox starts from an image; pass --image", ErrUnsupported)
	case len(cfg.Secrets) > 0:
		return plan{}, fmt.Errorf("%w: a credential would appear in an argument of the exec command", ErrUnsupported)
	case cfg.Provider != nil:
		return plan{}, fmt.Errorf("%w: a provider credential would appear in an argument of the exec command", ErrUnsupported)
	case len(cfg.BindPaths) > 0 || len(cfg.RoBindPaths) > 0 || len(c.RoBindPaths) > 0:
		return plan{}, fmt.Errorf("%w: a sandbox on a gateway has no host bind; the work directory is uploaded", ErrUnsupported)
	case len(c.PathEntries) > 0:
		return plan{}, fmt.Errorf("%w: a provisioned tool directory cannot reach the sandbox PATH", ErrUnsupported)
	}
	if strings.HasPrefix(cfg.Image, "-") {
		return plan{}, fmt.Errorf("%w: image %q starts with a hyphen", ErrUnsupported, cfg.Image)
	}
	policy, err := renderPolicy(cfg.Network, cfg.Egress.Allow)
	if err != nil {
		return plan{}, err
	}
	workDir, err := hostlaunch.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return plan{}, err
	}
	command := hostlaunch.WrapWithInitCommands(cfg.Process.Argv, c.InitCommands)
	env, err := hostlaunch.RunEnv(cfg)
	if err != nil {
		return plan{}, err
	}
	values := envutil.MergeEnvMaps(env, c.Env)

	remote := path.Join(sandboxRoot, filepath.Base(workDir))
	exec := []string{"sandbox", "exec", "--name", name, "--workdir", remote, "--no-tty"}
	names := make([]string, 0, len(values))
	for variable := range values {
		names = append(names, variable)
	}
	sort.Strings(names)
	for _, variable := range names {
		exec = append(exec, "--env", variable+"="+values[variable])
	}
	exec = append(exec, "--")
	exec = append(exec, command...)

	return plan{
		name:   name,
		policy: policy,
		create: func(policyFile string) []string {
			return []string{"sandbox", "create", "--name", name, "--from", cfg.Image, "--policy", policyFile, "--no-auto-providers", "--detach"}
		},
		upload:   []string{"sandbox", "upload", name, workDir, remote, "--no-git-ignore"},
		exec:     exec,
		download: []string{"sandbox", "download", name, remote, filepath.Dir(workDir)},
		remove:   []string{"sandbox", "delete", name},
	}, nil
}
