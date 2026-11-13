package phasetwo

import (
	"errors"
	"fmt"
	"net/http"
	"path"
)

// IDFromLocation returns the id of a resource the API just created: the last path
// segment of the response's Location header. Create operations return 201 with an
// empty body, so this is how to get the new id:
//
//	resp, err := client.CreateOrganizationWithResponse(ctx, realm, org)
//	...
//	id, err := phasetwo.IDFromLocation(resp.HTTPResponse)
func IDFromLocation(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errors.New("phasetwo: nil response")
	}
	loc, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("phasetwo: response has no Location header (status %d)", resp.StatusCode)
	}
	return path.Base(loc.Path), nil
}

// APIError is an error response from the Phase Two API.
type APIError struct {
	StatusCode int
	// Body is the raw response body; Keycloak usually sends {"error": "..."} JSON.
	Body []byte
}

func (e *APIError) Error() string {
	if len(e.Body) == 0 {
		return fmt.Sprintf("phasetwo: HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("phasetwo: HTTP %d: %s", e.StatusCode, e.Body)
}

// Check returns an *APIError if resp isn't a 2xx response. Pass the HTTPResponse
// and Body fields of any ...WithResponse result:
//
//	resp, err := client.GetOrganizationByIdWithResponse(ctx, realm, id)
//	if err != nil { ... }
//	if err := phasetwo.Check(resp.HTTPResponse, resp.Body); err != nil { ... }
func Check(resp *http.Response, body []byte) error {
	if resp == nil {
		return errors.New("phasetwo: nil response")
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return &APIError{StatusCode: resp.StatusCode, Body: body}
}
