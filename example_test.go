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

// Create an organization, make a link to its IdP Wizard, and send an audit event.
// This example is compiled but not run by `go test`.
func Example() {
	ctx := context.Background()
	realm := "my-realm"

	client, err := phasetwo.New(ctx, phasetwo.Config{
		BaseURL:      "https://auth.example.com/auth",
		TokenRealm:   realm,
		ClientID:     "api-client",
		ClientSecret: os.Getenv("PHASETWO_CLIENT_SECRET"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create an organization. The new id is in the Location header.
	name := "example"
	created, err := client.CreateOrganizationWithResponse(ctx, realm, openapi.OrganizationRepresentation{Name: &name})
	if err != nil {
		log.Fatal(err)
	}
	if err := phasetwo.Check(created.HTTPResponse, created.Body); err != nil {
		log.Fatal(err)
	}
	orgID, _ := phasetwo.IDFromLocation(created.HTTPResponse)

	// Create a link to the IdP Wizard for the organization
	link, err := client.CreatePortalLinkWithFormdataBodyWithResponse(ctx, realm, orgID, openapi.CreatePortalLinkFormdataRequestBody{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*link.JSON200.Link)

	// Send an audit event
	eventType := "foo.bar"
	now := time.Now().UnixMilli()
	if _, err := client.CreateEventWithResponse(ctx, realm, openapi.EventRepresentation{Type: &eventType, Time: &now}); err != nil {
		log.Fatal(err)
	}
}
