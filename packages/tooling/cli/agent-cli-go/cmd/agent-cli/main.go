package main

import (
	"os"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/cmd"
)

func main() {
	log := logging.FromEnvironment(os.Getenv, os.Stdout, os.Stderr)
	if err := cmd.NewRootCommand(log).Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
