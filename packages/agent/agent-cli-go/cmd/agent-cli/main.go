package main

import (
	"os"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/cmd"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/logging"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logging.LogError(err.Error())
		os.Exit(1)
	}
}
