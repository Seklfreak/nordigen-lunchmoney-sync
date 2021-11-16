package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Transaction represents a transaction.
type Transaction struct {
	// required parameters
	Date   string  `json:"date"` // TODO: make this a time.Time
	Amount float64 `json:"amount"`

	// optional parameters
	CategoryID  int      `json:"category_id,omitempty"`
	Payee       string   `json:"payee,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	AssetID     int      `json:"asset_id,omitempty"`
	RecurringID int      `json:"recurring_id,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Status      string   `json:"status,omitempty"`
	ExternalID  string   `json:"external_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// InsertTransactions inserts transactions to the Lunchmoney API.
func (c *Client) InsertTransactions(ctx context.Context, trx []*Transaction) (int, error) {
	request := struct {
		Transactions      []*Transaction `json:"transactions"`
		ApplyRules        bool           `json:"apply_rules"`
		SkipDuplicates    bool           `json:"skip_duplicates"`
		CheckForRecurring bool           `json:"check_for_recurring"`
		DebitAsNegative   bool           `json:"debit_as_negative"`
		SkipBalanceUpdate bool           `json:"skip_balance_update"`
	}{
		Transactions:      trx,
		ApplyRules:        true,
		SkipDuplicates:    true,
		CheckForRecurring: true,
		DebitAsNegative:   true,
		SkipBalanceUpdate: true,
	}

	reqData, err := json.Marshal(request)
	if err != nil {
		return 0, errors.Wrap(err, "failed to marshal request")
	}

	req, err := c.createRequest(ctx, http.MethodPost, "/v1/transactions", reqData)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("received unexpected status code when inserting transasction: %s", resp.Status)
	}

	var result struct {
		Error []string `json:"error"`
		IDs   []int    `json:"ids"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, errors.Wrap(err, "failed to decode response body")
	}

	if len(result.Error) > 0 {
		return len(result.IDs), fmt.Errorf("received %d errors: %q", len(result.Error), strings.Join(result.Error, "; "))
	}

	return len(result.IDs), nil
}
