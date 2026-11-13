// Package phasetwo is a Go client for the Phase Two Admin REST API: the
// organizations, SSO, webhooks, events and other extensions that Phase Two adds to
// Keycloak.
//
// The API operations and models are generated into the [openapi] package from the
// Phase Two OpenAPI spec. This package adds the hand-written parts: building an
// authenticated client ([New]), getting the id of a created resource
// ([IDFromLocation]) and turning error responses into errors ([Check]).
//
// For core Keycloak administration (realms, users, clients), use a Keycloak admin
// client alongside this one.
package phasetwo

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/p2-inc/phasetwo-go/openapi"
)

// Config configures a [Client].
type Config struct {
	// BaseURL is your Keycloak base URL, including the context path if it has one,
	// e.g. "https://auth.example.com/auth". Required.
	BaseURL string

	// TokenRealm, ClientID and ClientSecret get tokens with the client credentials
	// grant from TokenRealm's token endpoint, typically a service account in the
	// realm you manage. Ignored when TokenSource is set.
	TokenRealm   string
	ClientID     string
	ClientSecret string

	// TokenSource supplies access tokens instead of the client credentials grant,
	// e.g. a token from your own login flow.
	TokenSource oauth2.TokenSource

	// HTTPClient is used for both token and API requests (timeouts, proxies,
	// custom transports). Defaults to http.DefaultClient.
	HTTPClient *http.Client

	// UserAgent is sent on API requests. Defaults to "phasetwo-go".
	UserAgent string
}

// Client is an authenticated Phase Two API client. All API operations are
// methods of the embedded [openapi.ClientWithResponses]; every operation takes the
// realm name as its first argument after the context.
type Client struct {
	*openapi.ClientWithResponses
}

// New returns a Client that adds a bearer token to every request and refreshes it
// when it expires.
func New(ctx context.Context, cfg Config) (*Client, error) {
	base := strings.TrimRight(cfg.BaseURL, "/")
	if base == "" {
		return nil, errors.New("phasetwo: Config.BaseURL is required")
	}
	if _, err := url.Parse(base); err != nil {
		return nil, err
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	ts := cfg.TokenSource
	if ts == nil {
		if cfg.TokenRealm == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
			return nil, errors.New("phasetwo: set TokenSource, or TokenRealm, ClientID and ClientSecret")
		}
		cc := clientcredentials.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			TokenURL:     base + "/realms/" + url.PathEscape(cfg.TokenRealm) + "/protocol/openid-connect/token",
		}
		ts = cc.TokenSource(context.WithValue(ctx, oauth2.HTTPClient, hc))
	}

	authed := *hc
	authed.Transport = &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, ts), Base: hc.Transport}

	ua := cfg.UserAgent
	if ua == "" {
		ua = "phasetwo-go"
	}
	c, err := openapi.NewClientWithResponses(base+"/realms",
		openapi.WithHTTPClient(&authed),
		openapi.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("User-Agent", ua)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}
