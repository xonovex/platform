module github.com/xonovex/platform/packages/shared/shared-agent-go

go 1.25.5

require (
	github.com/pelletier/go-toml/v2 v2.2.4
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/xonovex/platform/packages/shared/shared-core-go => ../shared-core-go
