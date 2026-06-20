module github.com/xonovex/platform/packages/cli/agent-cli-go

go 1.26.0

require (
	github.com/spf13/cobra v1.10.2
	github.com/xonovex/platform/packages/shared/shared-agent-go v0.0.0-20260613164631-f8286f3d1667
	github.com/xonovex/platform/packages/shared/shared-core-go v0.0.0-20260613164631-f8286f3d1667
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/xonovex/platform/packages/shared/shared-agent-go => ../../shared/shared-agent-go

replace github.com/xonovex/platform/packages/shared/shared-core-go => ../../shared/shared-core-go
