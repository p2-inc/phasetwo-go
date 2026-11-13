//go:build integration

// Integration test against a real Phase Two Keycloak in Docker:
//
//	go test -tags integration ./...
package phasetwo_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/oauth2"

	"github.com/p2-inc/phasetwo-go"
	"github.com/p2-inc/phasetwo-go/openapi"
)

const keycloakImage = "quay.io/phasetwo/phasetwo-keycloak:26.5.6"

func startKeycloak(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := testcontainers.Run(ctx, keycloakImage,
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithEnv(map[string]string{
			"KC_BOOTSTRAP_ADMIN_USERNAME": "admin",
			"KC_BOOTSTRAP_ADMIN_PASSWORD": "admin",
			"KC_HTTP_RELATIVE_PATH":       "/auth",
		}),
		testcontainers.WithCmd("start-dev"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/auth/realms/master").WithPort("8080/tcp").WithStartupTimeout(3*time.Minute)),
	)
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := c.PortEndpoint(ctx, "8080/tcp", "http")
	if err != nil {
		t.Fatal(err)
	}
	return endpoint + "/auth"
}

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	base := startKeycloak(t)

	// The bootstrap admin uses the password grant; real callers would use a
	// service account and the client credentials grant via Config.
	oc := oauth2.Config{ClientID: "admin-cli", Endpoint: oauth2.Endpoint{
		TokenURL: base + "/realms/master/protocol/openid-connect/token"}}
	tok, err := oc.PasswordCredentialsToken(ctx, "admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	client, err := phasetwo.New(ctx, phasetwo.Config{BaseURL: base, TokenSource: oauth2.StaticTokenSource(tok)})
	if err != nil {
		t.Fatal(err)
	}
	realm := "master"

	// The README quick start (example_test.go), with a new user as the member.
	memberID := createUser(t, ctx, oc.Client(ctx, tok), base, realm, fmt.Sprintf("member-%d", time.Now().UnixNano()))
	qsOrg, err := quickstart(ctx, client, realm, memberID)
	if err != nil {
		t.Fatalf("quickstart: %v", err)
	}
	isMember, err := client.CheckOrganizationMembershipWithResponse(ctx, realm, qsOrg, memberID)
	must(t, isMember.HTTPResponse, isMember.Body, err)
	qsDel, err := client.DeleteOrganizationWithResponse(ctx, realm, qsOrg)
	must(t, qsDel.HTTPResponse, qsDel.Body, err)

	name := fmt.Sprintf("go-%d", time.Now().UnixNano())
	created, err := client.CreateOrganizationWithResponse(ctx, realm, openapi.OrganizationRepresentation{Name: &name})
	must(t, created.HTTPResponse, created.Body, err)
	orgID, err := phasetwo.IDFromLocation(created.HTTPResponse)
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.GetOrganizationByIdWithResponse(ctx, realm, orgID)
	must(t, got.HTTPResponse, got.Body, err)
	if *got.JSON200.Name != name {
		t.Fatalf("name = %q, want %q", *got.JSON200.Name, name)
	}

	list, err := client.GetOrganizationsWithResponse(ctx, realm, &openapi.GetOrganizationsParams{Search: &name})
	must(t, list.HTTPResponse, list.Body, err)
	if len(*list.JSON200) != 1 {
		t.Fatalf("search returned %d orgs", len(*list.JSON200))
	}

	link, err := client.CreatePortalLinkWithFormdataBodyWithResponse(ctx, realm, orgID, openapi.CreatePortalLinkFormdataRequestBody{})
	must(t, link.HTTPResponse, link.Body, err)
	if link.JSON200 == nil || link.JSON200.Link == nil || *link.JSON200.Link == "" {
		t.Fatalf("no portal link in %s", link.Body)
	}

	count, err := client.GetOrganizationMembershipsCountWithResponse(ctx, realm, orgID, nil)
	must(t, count.HTTPResponse, count.Body, err)
	t.Logf("org %s has %d member(s)", orgID, *count.JSON200)

	eventType, now := "foo.bar", time.Now().UnixMilli()
	ev, err := client.CreateEventWithResponse(ctx, realm, openapi.EventRepresentation{Type: &eventType, Time: &now})
	must(t, ev.HTTPResponse, ev.Body, err)

	del, err := client.DeleteOrganizationWithResponse(ctx, realm, orgID)
	must(t, del.HTTPResponse, del.Body, err)

	missing, err := client.GetOrganizationByIdWithResponse(ctx, realm, orgID)
	if err != nil {
		t.Fatal(err)
	}
	if err := phasetwo.Check(missing.HTTPResponse, missing.Body); err == nil {
		t.Fatal("expected an error for a deleted organization")
	}
}

func must(t *testing.T, resp *http.Response, body []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if err := phasetwo.Check(resp, body); err != nil {
		t.Fatal(err)
	}
}

// createUser creates a user with the Keycloak admin API (not part of this SDK)
// and returns its id.
func createUser(t *testing.T, ctx context.Context, hc *http.Client, base, realm, username string) string {
	t.Helper()
	body := strings.NewReader(fmt.Sprintf(`{"username":%q,"enabled":true}`, username))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/admin/realms/"+realm+"/users", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("creating user %q: %s", username, resp.Status)
	}
	id, err := phasetwo.IDFromLocation(resp)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
