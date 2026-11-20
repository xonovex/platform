module github.com/xonovex/platform/packages/tooling/cli/agent-cli-go

go 1.27.0

require (
	github.com/pelletier/go-toml/v2 v2.4.3
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/spf13/cobra v1.10.2
	github.com/xonovex/platform/packages/library/shared-agent-go v0.0.0-20261119162506-a6d3fcd2ee9a
	github.com/xonovex/platform/packages/library/shared-config-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-contract-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-core-go v0.0.0-20261119162506-a6d3fcd2ee9a
	github.com/xonovex/platform/packages/library/shared-launch-go v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.1.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/xonovex/platform/packages/library/shared-agent-go => ../../../library/shared-agent-go

replace github.com/xonovex/platform/packages/library/shared-config-go => ../../../library/shared-config-go

replace github.com/xonovex/platform/packages/library/shared-core-go => ../../../library/shared-core-go

replace github.com/xonovex/platform/packages/library/shared-contract-go => ../../../library/shared-contract-go

replace github.com/xonovex/platform/packages/library/shared-launch-go => ../../../library/shared-launch-go
