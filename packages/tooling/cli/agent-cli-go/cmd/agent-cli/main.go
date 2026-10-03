package main

import (
	"os"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logging.LogError(err.Error())
		os.Exit(1)
	}
}
