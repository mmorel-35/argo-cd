package github_app_auth

import "context"

// Authentication has the authentication information required to access the GitHub API and repositories.
type Authentication struct {
	// ID specifies the ID of the GitHub app used to access the repo
	ID int64
	// InstallationID specifies the installation ID of the GitHub App used to access the repo
	InstallationID int64
	// EnterpriseBaseURL specifies the base URL of GitHub Enterprise installation. If empty will default to https://api.github.com
	EnterpriseBaseURL string
	// PrivateKey in PEM format.
	PrivateKey string
}

type Credentials interface {
	GetAuthSecret(ctx context.Context, secretName string) (*Authentication, error)
}
