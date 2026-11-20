module github.com/xonovex/platform/packages/library/shared-agent-go

go 1.27.0

replace github.com/xonovex/platform/packages/library/shared-contract-go => ../shared-contract-go

require (
	github.com/xonovex/platform/packages/library/shared-contract-go v0.0.0-00010101000000-000000000000
	github.com/xonovex/platform/packages/library/shared-core-go v0.0.0-00010101000000-000000000000
)

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/xonovex/platform/packages/library/shared-core-go => ../shared-core-go
