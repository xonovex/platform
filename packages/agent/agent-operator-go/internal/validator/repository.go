package validator

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"

var (
	ValidateRepositoryURL = validation.ValidateRepositoryURL
	ValidateBranch        = validation.ValidateBranch
	ValidateCommit        = validation.ValidateCommit
)
