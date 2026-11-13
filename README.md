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

A Go client for the [Phase Two Admin REST API](https://phasetwo.io/api/phase-two-admin-rest-api): the organizations, members, roles, domains, identity providers, invitations, SCIM, webhooks, events and realm attributes that Phase Two adds to Keycloak. For core Keycloak administration (realms, users, clients), use a Keycloak admin client alongside it.

> [!NOTE]
> All Phase Two SDKs (Java, TypeScript, Python, Go) share the version of the Phase Two API they implement: every SDK release X.Y.* implements API X.Y, and patch releases are SDK-only fixes. While on 0.x, a new minor version can contain breaking changes from the API spec, so pin a version.

## Installation

```sh
go get github.com/p2-inc/phasetwo-go
```

Requires Go 1.26 or later.

## Quick start

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
```

The same code is in [`example_test.go`](example_test.go), which `go test` compiles.

## Authentication

`phasetwo.New` gets tokens with the client credentials grant and refreshes them as they expire. Create a confidential client with a [service account](https://phasetwo.io/docs/api/service-accounts) that has the roles your calls need.

To bring your own tokens (e.g. from a login flow), set `Config.TokenSource` to any `oauth2.TokenSource`. `Config.HTTPClient` sets timeouts, proxies or a custom transport for both token and API requests.

## Usage

Every operation in the [API reference](https://phasetwo.io/api/phase-two-admin-rest-api) is a method on `Client`, named after its operationId, with the realm as the first argument after the context:

- `XxxWithResponse(...)` returns a typed response: `JSON200` (or `JSON201`, ...) holds the decoded body, `HTTPResponse` and `Body` the raw response.
- Operations with request bodies take a typed body (`openapi.OrganizationRepresentation`, ...). There's also a `...WithBodyWithResponse` variant that takes an `io.Reader`.
- Optional query parameters are in a `*openapi.XxxParams` struct. Pass `nil` for none.

### Errors

A non-2xx response is not a Go error; check it with `phasetwo.Check(resp.HTTPResponse, resp.Body)`, which returns a `*phasetwo.APIError` with the status code and body.

## API reference

- Go package docs: [pkg.go.dev/github.com/p2-inc/phasetwo-go](https://pkg.go.dev/github.com/p2-inc/phasetwo-go)
- REST API: [phasetwo.io/api](https://phasetwo.io/api/phase-two-admin-rest-api)

## Generated code

The `openapi` package is generated from the [Phase Two OpenAPI spec](https://github.com/p2-inc/phasetwo-docs/blob/main/openapi.yaml) with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) by [phasetwo-sdk-generator](https://github.com/p2-inc/phasetwo-sdk-generator). Don't edit it by hand; `sdk-generator.json` records the spec commit it came from. Everything in the root `phasetwo` package is hand-written.

```sh
go test ./...                    # unit tests
go test -tags integration ./...  # against Phase Two Keycloak in Docker
```

## License

Apache License 2.0, see [COPYING](COPYING).
