package nordigen

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents a Nordigen API error.
type APIError struct {
	Summary    string `json:"summary"`
	Detail     string `json:"detail"`
	StatusCode int    `json:"status_code"`
}

// Error returns the error message of a Nordigen API error.
func (e APIError) Error() string {
	return fmt.Sprintf("received unexpected status code %d: %s (%s)", e.StatusCode, e.Summary, e.Detail)
}

// extractError extracts an APIError from a response if it is possible.
func extractError(resp *http.Response) error {
	var apiError APIError
	err := json.NewDecoder(resp.Body).Decode(&apiError)
	if err != nil {
		return nil
	}

	if apiError.StatusCode == 0 ||
		apiError.Summary == "" ||
		apiError.Detail == "" {
		return nil
	}

	return &apiError
}
