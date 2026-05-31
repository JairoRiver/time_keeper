package logto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/identity"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// Client is the Logto implementation of identity.Provider.
// It manages the OIDC authorization flow and validates id_tokens via JWKS.
type Client struct {
	endpoint    string
	appID       string
	appSecret   string
	callbackURL string
	jwksCache   *jwk.Cache
}

// Compile-time check: Client must satisfy identity.Provider.
var _ identity.Provider = (*Client)(nil)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// New creates a Logto client and registers the JWKS endpoint in the cache.
// WithWaitReady(true) makes it block until the first fetch succeeds, so the
// server fails fast at startup if Logto is unreachable.
func New(ctx context.Context, cfg util.Config) (*Client, error) {
	jwksURL := cfg.Logto.Endpoint + "/oidc/jwks"

	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("logto: create jwks cache: %w", err)
	}

	if err := cache.Register(ctx, jwksURL,
		jwk.WithMinInterval(15*time.Minute),
		jwk.WithWaitReady(true),
	); err != nil {
		return nil, fmt.Errorf("logto: register jwks url: %w", err)
	}

	return &Client{
		endpoint:    cfg.Logto.Endpoint,
		appID:       cfg.Logto.AppID,
		appSecret:   cfg.Logto.AppSecret,
		callbackURL: cfg.Logto.CallbackURL,
		jwksCache:   cache,
	}, nil
}

// BuildAuthURL returns the Logto authorization URL.
func (c *Client) BuildAuthURL(state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.appID)
	params.Set("redirect_uri", c.callbackURL)
	params.Set("scope", "openid profile email")
	params.Set("state", state)
	return c.endpoint + "/oidc/auth?" + params.Encode()
}

// ExchangeCode calls Logto's token endpoint and validates the returned id_token.
func (c *Client) ExchangeCode(ctx context.Context, code string) (*identity.UserClaims, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.callbackURL)
	form.Set("client_id", c.appID)
	form.Set("client_secret", c.appSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint+"/oidc/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("logto: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("logto: token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("logto: token endpoint returned %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("logto: decode token response: %w", err)
	}

	return c.validateIDToken(ctx, tr.IDToken)
}

// validateIDToken verifies the JWT signature against the cached JWKS and
// extracts the subject and email claims.
func (c *Client) validateIDToken(ctx context.Context, idToken string) (*identity.UserClaims, error) {
	keySet, err := c.jwksCache.Lookup(ctx, c.endpoint+"/oidc/jwks")
	if err != nil {
		return nil, fmt.Errorf("logto: lookup jwks: %w", err)
	}

	tok, err := jwt.Parse([]byte(idToken),
		jwt.WithKeySet(keySet),
		jwt.WithIssuer(c.endpoint+"/oidc"),
	)
	if err != nil {
		return nil, fmt.Errorf("logto: validate id_token: %w", err)
	}

	sub, ok := tok.Subject()
	if !ok {
		return nil, fmt.Errorf("logto: id_token missing subject claim")
	}

	claims := &identity.UserClaims{Sub: sub}
	var email string
	if err := tok.Get("email", &email); err == nil {
		claims.Email = email
	}

	return claims, nil
}
