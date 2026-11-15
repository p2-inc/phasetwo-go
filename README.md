# Phase Two Go SDK

<!-- phasetwo:banner:start - managed by p2-inc/shared-github-actions, edit there -->
[![build](https://img.shields.io/github/actions/workflow/status/p2-inc/phasetwo-go/ci.yml?branch=main&label=build)](https://github.com/p2-inc/phasetwo-go/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/p2-inc/phasetwo-go.svg)](https://pkg.go.dev/github.com/p2-inc/phasetwo-go) [![license](https://img.shields.io/badge/license-Apache--2.0-blue)](COPYING) [![docs](https://img.shields.io/badge/docs-phasetwo.io-blue)](https://phasetwo.io/docs/api/sdks/)

📚 [Documentation](https://phasetwo.io/docs/api/sdks/) • 🚀 [Hosted Keycloak](https://dash.phasetwo.io/?utm_source=github&utm_medium=readme&utm_campaign=phasetwo-go) • 🤖 [Keycloak Skills](https://github.com/p2-inc/keycloak-skills) • 💬 [Support](mailto:support@phasetwo.io)

> [!TIP]
> **Run this in our cloud, no setup needed.** Phase Two hosts Keycloak with all of our extensions built in. [Try it free with your own cluster →](https://dash.phasetwo.io/?utm_source=github&utm_medium=readme&utm_campaign=phasetwo-go)
>
> **Using a coding agent?** Add the Phase Two Keycloak skills ([launch post](https://phasetwo.io/blog/keycloak-skills-mcp-launch)) to Claude Code, Cursor or Codex so your agent can set up passwordless login, SSO, organizations and clusters for you:
>
> ```sh
> # Install the skills (connects to our hosted MCP server)
> claude plugin marketplace add p2-inc/keycloak-skills
> claude plugin install phasetwo@keycloak-skills
> # Or connect to the MCP server directly
> claude mcp add --transport http keycloak https://mcp.phasetwo.io/mcp
> ```
<!-- phasetwo:banner:end -->

A Go client for the [Phase Two Admin REST API](https://phasetwo.io/api/phase-two-admin-rest-api): organizations, members, roles, domains, identity providers, invitations, SCIM, webhooks and events, realm attributes and magic links, all added to Keycloak by Phase Two. For core Keycloak administration (realms, users, clients), use a Keycloak admin client such as [gocloak](https://github.com/Nerzal/gocloak) alongside it.

> [!NOTE]
> Pre-1.0: minor versions can include breaking changes. All Phase Two SDKs share a version: release X.Y.* implements Phase Two API X.Y.

## Installation

```sh
go get github.com/p2-inc/phasetwo-go
```

Requires Go 1.26 or later.

## Quick start

Connect with a service account:

<!-- phasetwo:snippet client example_test.go -->
```go
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
```
<!-- /phasetwo:snippet -->

Create an organization, add a member, create an IdP Wizard link so the organization's admin can set up SSO, and send an audit event:

<!-- phasetwo:snippet quickstart example_test.go -->
```go
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
```
<!-- /phasetwo:snippet -->

This code is in [`example_test.go`](example_test.go). The integration test runs it against Phase Two Keycloak in CI.

## Authentication

`phasetwo.New` gets tokens with the client credentials grant and refreshes them as they expire. Create a confidential client with a [service account](https://phasetwo.io/docs/api/service-accounts) that has the roles your calls need, and pass its client id and secret in `Config`.

To bring your own tokens (e.g. from a login flow), set `Config.TokenSource` to any `oauth2.TokenSource`. `Config.HTTPClient` sets timeouts, proxies or a custom transport for both token and API requests.

## Usage

- Every operation in the [API reference](https://phasetwo.io/api/phase-two-admin-rest-api) is a method on `Client`, named after its operationId, with the realm as the first argument after the context.
- `XxxWithResponse(...)` returns a typed response: `JSON200` (or `JSON201`, ...) holds the decoded body, and `HTTPResponse` and `Body` hold the raw response.
- Operations with a request body take a typed body (`openapi.OrganizationRepresentation`, ...). The `...WithBodyWithResponse` variant takes an `io.Reader` instead.
- Optional query parameters are in an `*openapi.XxxParams` struct. Pass `nil` for none.
- Creating a resource returns `201 Created` with an empty body. Get the new id from the `Location` header with `phasetwo.IDFromLocation(resp.HTTPResponse)`.

### Errors

A non-2xx response is not a Go error. Check it with `phasetwo.Check(resp.HTTPResponse, resp.Body)`, which returns a `*phasetwo.APIError` with the status code and body:

```go
var apiErr *phasetwo.APIError
if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
	// ...
}
```

## Compatibility

| SDK | Phase Two API | Tested against |
|---|---|---|
| v0.2.x | 0.2 | `quay.io/phasetwo/phasetwo-keycloak:26.5.6` |

The API version and the spec commit the SDK was generated from are in [`sdk-generator.json`](sdk-generator.json). CI tests the two latest Go releases, matching Go's own support policy.

## API reference

- Go package docs: [pkg.go.dev/github.com/p2-inc/phasetwo-go](https://pkg.go.dev/github.com/p2-inc/phasetwo-go)
- REST API: [phasetwo.io/api](https://phasetwo.io/api/phase-two-admin-rest-api)

## Generated code

The `openapi` package is generated from the [Phase Two OpenAPI spec](https://github.com/p2-inc/phasetwo-docs/blob/main/openapi.yaml) with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) by [phasetwo-sdk-generator](https://github.com/p2-inc/phasetwo-sdk-generator). Don't edit it by hand: CI regenerates it and fails if it differs. Everything in the root `phasetwo` package is hand-written.

```sh
go test ./...                    # unit tests
go test -tags integration ./...  # against Phase Two Keycloak in Docker
```

## License

Apache License 2.0, see [COPYING](COPYING).
