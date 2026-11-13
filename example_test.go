package phasetwo_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/p2-inc/phasetwo-go"
	"github.com/p2-inc/phasetwo-go/openapi"
)

// Connect with a service account, then run the quick start. This example is
// compiled but not run by `go test`; the integration test runs quickstart
// against Phase Two Keycloak.
func Example() {
	// snippet:start client
	ctx := context.Background()
	realm := "my-realm"

	client, err := phasetwo.New(ctx, phasetwo.Config{
		BaseURL:      "https://auth.example.com/auth", // your Keycloak base URL
		TokenRealm:   realm,                           // where the service account lives
		ClientID:     "api-client",
		ClientSecret: os.Getenv("PHASETWO_CLIENT_SECRET"),
	})
	if err != nil {
		log.Fatal(err)
	}
	// snippet:end client

	if _, err := quickstart(ctx, client, realm, "a-user-id"); err != nil {
		log.Fatal(err)
	}
}

// quickstart creates an organization, adds a member, makes an IdP Wizard link
// for the organization and sends an audit event. It returns the organization id.
func quickstart(ctx context.Context, client *phasetwo.Client, realm, userID string) (string, error) {
	// snippet:start quickstart
	// Create an organization. The new id is in the Location header.
	name := "example"
	created, err := client.CreateOrganizationWithResponse(ctx, realm,
		openapi.OrganizationRepresentation{Name: &name})
	if err != nil {
		return "", err
	}
	if err := phasetwo.Check(created.HTTPResponse, created.Body); err != nil {
		return "", err
	}
	orgID, err := phasetwo.IDFromLocation(created.HTTPResponse)
	if err != nil {
		return "", err
	}

	// Add an existing user as a member.
	added, err := client.AddOrganizationMemberWithResponse(ctx, realm, orgID, userID)
	if err != nil {
		return "", err
	}
	if err := phasetwo.Check(added.HTTPResponse, added.Body); err != nil {
		return "", err
	}

	// Create a link to the IdP Wizard, so the organization's admin can set up SSO.
	link, err := client.CreatePortalLinkWithFormdataBodyWithResponse(ctx, realm, orgID,
		openapi.CreatePortalLinkFormdataRequestBody{})
	if err != nil {
		return "", err
	}
	if err := phasetwo.Check(link.HTTPResponse, link.Body); err != nil {
		return "", err
	}
	fmt.Println("IdP Wizard:", *link.JSON200.Link)

	// Send an audit event.
	eventType := "example.created"
	now := time.Now().UnixMilli()
	sent, err := client.CreateEventWithResponse(ctx, realm,
		openapi.EventRepresentation{Type: &eventType, Time: &now})
	if err != nil {
		return "", err
	}
	if err := phasetwo.Check(sent.HTTPResponse, sent.Body); err != nil {
		return "", err
	}
	// snippet:end quickstart
	return orgID, nil
}
