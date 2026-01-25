package main

import (
	"os"

	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/cmd"
	"github.com/xonovex/platform/packages/tools/tool-lib-go/pkg/logging"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logging.LogError(err.Error())
		os.Exit(1)
	}
}
