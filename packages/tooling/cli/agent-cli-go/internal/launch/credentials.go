package launch

import (
	"errors"

	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
)

// credentialValues resolves the credential bindings of the run. A dry run
// resolves nothing, so it prompts no keychain and holds no secret: it carries the
// redacted placeholder of each credential, which is enough to build the command.
func credentialValues(registry *credshared.Registry, plan Plan) (map[string]string, error) {
	if len(plan.Credentials) == 0 {
		return nil, nil
	}
	if plan.DryRun {
		placeholders := make(map[string]string, len(plan.Credentials))
		for _, binding := range plan.Credentials {
			placeholders[binding.Name] = credshared.Redacted
		}
		return placeholders, nil
	}
	if registry == nil {
		return nil, errors.New("this build has no credential stores")
	}
	secrets, err := credshared.Resolve(registry, plan.Credentials)
	if err != nil {
		return nil, err
	}
	return credshared.Reveal(secrets), nil
}
