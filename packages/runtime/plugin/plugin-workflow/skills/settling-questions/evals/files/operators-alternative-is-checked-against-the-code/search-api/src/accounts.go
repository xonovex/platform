package accounts

type AccountAPI interface {
	GetProfile(accountID string) (Profile, error)
	UpdateProfile(accountID string, patch ProfilePatch) (Profile, error)
}

// No bulk upload: every write is a single-profile patch.
