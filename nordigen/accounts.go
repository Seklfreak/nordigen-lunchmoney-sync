package nordigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Account represents an account.
type Account struct {
	ResourceID      string `json:"resourceId"`
	Currency        string `json:"currency"`
	Name            string `json:"name"`
	OwnerName       string `json:"ownerName"`
	Product         string `json:"product"`
	CashAccountType string `json:"cashAccountType"`
	Status          string `json:"status"`
	IBAN            string `json:"iban"`
	BIC             string `json:"bic"`
	Usage           string `json:"usage"`
}

// GetAccountDetails fetches details for an account.
func (c *Client) GetAccountDetails(ctx context.Context, accountID string) (*Account, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/accounts/%s/details/", baseURL, accountID),
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
		if err := extractError(resp); err != nil {
			return nil, errors.Wrap(err, "failed to fetch account details")
		}

		return nil, errors.Errorf("received unexpected status code when fetching account details: %s", resp.Status)
	}

	var account struct {
		Account *Account `json:"account"`
	}

	err = json.NewDecoder(resp.Body).Decode(&account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return account.Account, nil
}

// Balance represents an account balance.
type Balance struct {
	BalanceAmount      Amount    `json:"balanceAmount"`
	BalanceType        string    `json:"balanceType"`
	LastChangeDateTime time.Time `json:"lastChangeDateTime"`
	ReferenceDate      Date      `json:"referenceDate"`
}

// GetAccountBalances fetches the balances for an account.
func (c *Client) GetAccountBalances(ctx context.Context, accountID string) ([]*Balance, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/accounts/%s/balances/", baseURL, accountID),
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
		if err := extractError(resp); err != nil {
			return nil, errors.Wrap(err, "failed to fetch account balances")
		}

		return nil, errors.Errorf("received unexpected status code when fetching account balances: %s", resp.Status)
	}

	var balances struct {
		Balances []*Balance `json:"balances"`
	}

	err = json.NewDecoder(resp.Body).Decode(&balances)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return balances.Balances, nil
}
