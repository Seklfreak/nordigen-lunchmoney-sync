package nordigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// AccountList represents a list of accounts.
type AccountList struct {
	ID               string    `json:"id"`
	Created          time.Time `json:"created"`
	Redirect         string    `json:"redirect"`
	Status           string    `json:"status"`
	InstitutionID    string    `json:"institution_id"`
	Agreement        string    `json:"agreement"`
	Reference        string    `json:"reference"`
	Accounts         []string  `json:"accounts"`
	UserLanguage     string    `json:"user_language"`
	Link             string    `json:"link"`
	AccountSelection bool      `json:"account_selection"`
}

// ListAccounts fetches Nordigen accounts.
func (c *Client) ListAccounts(ctx context.Context, requisitionID string) (*AccountList, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/requisitions/%s/", baseURL, requisitionID),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Bearer "+c.accessKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received unexpected status code when fetching accounts: %s", resp.Status)
	}

	var accountList AccountList

	err = json.NewDecoder(resp.Body).Decode(&accountList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return &accountList, nil
}
