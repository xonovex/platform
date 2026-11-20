package managed

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	isohost "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
)

// BridgeHost is the port the bridge consumes. It runs inside the sandbox.
type BridgeHost interface {
	Listen(network, address string) (net.Listener, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Environ() []string
	Run(name string, args, env []string, dir string) (int, error)
}

// ErrNoCommand reports a bridge started without a command to run.
var ErrNoCommand = errors.New("egress bridge needs a command to run")

// Bridge relays connections to the loopback address of a sandbox that has no
// network to the proxy socket that the host bound in, and runs one command whose
// proxy environment points at it.
type Bridge struct{ Host BridgeHost }

// Run listens on listen, relays each connection to socket, runs command with the
// proxy environment, and returns the exit code of command. The proxy variables
// replace any the command would inherit, so the agent cannot be steered around the
// proxy by the environment it started in.
func (b Bridge) Run(socket, listen string, command []string) (int, error) {
	if len(command) == 0 {
		return 1, ErrNoCommand
	}
	listener, err := b.Host.Listen("tcp", listen)
	if err != nil {
		return 1, fmt.Errorf("listen on %s: %w", listen, err)
	}
	defer func() { _ = listener.Close() }()
	go b.serve(listener, socket)
	env := envutil.MergeEnvMaps(envutil.ParseEnv(b.Host.Environ()), isohost.ProxyEnv(listen))
	code, err := b.Host.Run(command[0], command[1:], envutil.EnvMapToSlice(env), "")
	if err != nil {
		return 1, fmt.Errorf("run %s: %w", command[0], err)
	}
	return code, nil
}

func (b Bridge) serve(listener net.Listener, socket string) {
	for {
		client, err := listener.Accept()
		if err != nil {
			return
		}
		go func() {
			upstream, err := b.Host.DialContext(context.Background(), "unix", socket)
			if err != nil {
				_ = client.Close()
				return
			}
			relay(client, client, upstream)
		}()
	}
}
