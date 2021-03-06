package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Asset represents a single asset.
type Asset struct {
	ID        int        `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// parameters to update.
	TypeName        string        `json:"type_name,omitempty"`
	SubtypeName     string        `json:"subtype_name,omitempty"`
	Name            string        `json:"name,omitempty"`
	Balance         *AssetBalance `json:"balance,omitempty"`
	BalanceAsOf     *time.Time    `json:"balance_as_of,omitempty"`
	Currency        string        `json:"currency,omitempty"`
	InstitutionName string        `json:"institution_name,omitempty"`
}

// AssetBalance represents the balance of an asset.
type AssetBalance float64

// UnmarshalJSON provides custom JSON unmarshalling for AssetBalance.
func (ab *AssetBalance) UnmarshalJSON(b []byte) (err error) {
	amount, err := strconv.ParseFloat(strings.Trim(string(b), "\""), 64)
	if err != nil {
		return errors.Wrapf(err, "could not parse transaction amount value %q", strings.Trim(string(b), "\""))
	}

	*ab = AssetBalance(amount)

	return nil
}

// MarshalJSON provides custom JSON marshalling for AssetBalance.
func (ab *AssetBalance) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%.2f\"", float64(*ab))), nil
}

// GetAssets retrieves assets from the Lunchmoney API.
func (c *Client) GetAssets(ctx context.Context) ([]*Asset, error) {
	req, err := c.createRequest(ctx, http.MethodGet, "/v1/assets", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received unexpected status code when fetching assets: %s", resp.Status)
	}

	var assetsContainer struct {
		Assets []*Asset `json:"assets"`
	}

	err = json.NewDecoder(resp.Body).Decode(&assetsContainer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return assetsContainer.Assets, nil
}

// UpdateAsset updates an asset in the Lunchmoney API.
func (c *Client) UpdateAsset(ctx context.Context, assetID int, asset *Asset) error {
	reqData, err := json.Marshal(asset)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request")
	}

	req, err := c.createRequest(ctx, http.MethodPut, fmt.Sprintf("/v1/assets/%d", assetID), reqData)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code when updating asset: %s", resp.Status)
	}

	var result struct {
		Errors []string `json:"errors"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("received %d errors: %q", len(result.Errors), strings.Join(result.Errors, "; "))
	}

	return nil
}
