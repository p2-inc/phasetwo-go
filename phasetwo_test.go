package phasetwo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p2-inc/phasetwo-go"
	"github.com/p2-inc/phasetwo-go/openapi"
)

// A fake Keycloak: a token endpoint and the create-organization endpoint.
func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/realms/svc/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.Form.Get("grant_type") != "client_credentials" {
			http.Error(w, "bad grant", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-1","token_type":"Bearer","expires_in":300}`))
	})
	mux.HandleFunc("POST /auth/realms/acme/orgs", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok-1" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if r.Header.Get("User-Agent") != "phasetwo-go" {
			http.Error(w, "missing user agent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", "http://"+r.Host+"/auth/realms/acme/orgs/org-123")
		w.WriteHeader(http.StatusCreated)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestCreateOrganizationWithClientCredentials(t *testing.T) {
	srv := fakeServer(t)
	ctx := context.Background()
	c, err := phasetwo.New(ctx, phasetwo.Config{
		BaseURL: srv.URL + "/auth/", TokenRealm: "svc", ClientID: "api", ClientSecret: "s3cret",
	})
	if err != nil {
		t.Fatal(err)
	}
	name := "example"
	resp, err := c.CreateOrganizationWithResponse(ctx, "acme", openapi.OrganizationRepresentation{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if err := phasetwo.Check(resp.HTTPResponse, resp.Body); err != nil {
		t.Fatal(err)
	}
	id, err := phasetwo.IDFromLocation(resp.HTTPResponse)
	if err != nil || id != "org-123" {
		t.Fatalf("IDFromLocation = %q, %v", id, err)
	}
}

func TestCheckReturnsAPIError(t *testing.T) {
	err := phasetwo.Check(&http.Response{StatusCode: 404}, []byte(`{"error":"not found"}`))
	var apiErr *phasetwo.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Fatalf("got %v", err)
	}
}

func TestNewRequiresCredentials(t *testing.T) {
	if _, err := phasetwo.New(context.Background(), phasetwo.Config{BaseURL: "https://example.com/auth"}); err == nil {
		t.Fatal("expected an error without credentials")
	}
}
