module github.com/xonovex/platform/packages/tools/tool-agent-cli-go

go 1.25.5

require (
	github.com/xonovex/platform/packages/tools/tool-lib-go v0.0.0-00010101000000-000000000000
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/spf13/cobra v1.8.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/xonovex/platform/packages/tools/tool-lib-go => ../tool-lib-go
