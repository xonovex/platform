module github.com/xonovex/platform/packages/tooling/cli/agent-eval-go

go 1.27.0

replace github.com/xonovex/platform/packages/library/shared-config-go => ../../../library/shared-config-go

replace github.com/xonovex/platform/packages/library/shared-contract-go => ../../../library/shared-contract-go

replace github.com/xonovex/platform/packages/library/shared-core-go => ../../../library/shared-core-go

replace github.com/xonovex/platform/packages/library/shared-launch-go => ../../../library/shared-launch-go

require (
	github.com/spf13/cobra v1.10.2
	github.com/xonovex/platform/packages/library/shared-agent-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-config-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-contract-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-core-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-launch-go v0.0.0-00010101000000-000000000000
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/xonovex/platform/packages/library/shared-agent-go => ../../../library/shared-agent-go
