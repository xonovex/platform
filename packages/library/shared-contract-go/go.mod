module github.com/xonovex/platform/packages/library/shared-contract-go

go 1.27.0

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/xonovex/platform/packages/library/shared-core-go v0.0.0-00010101000000-000000000000
)

require golang.org/x/text v0.40.0 // indirect

replace github.com/xonovex/platform/packages/library/shared-core-go => ../shared-core-go
