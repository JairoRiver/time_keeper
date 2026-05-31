package identity

import "context"

// UserClaims holds the verified identity returned by any provider.
// Sub is the provider's stable user ID — stored as user_identity_id internally.
type UserClaims struct {
	Sub   string
	Email string
}

// Provider is the interface any identity provider must implement.
type Provider interface {
	// BuildAuthURL returns the authorization URL to redirect the user to.
	// state is a random nonce the caller generates for CSRF protection.
	BuildAuthURL(state string) string

	// ExchangeCode exchanges an authorization code (from the OIDC callback)
	// for verified user claims.
	ExchangeCode(ctx context.Context, code string) (*UserClaims, error)
}
