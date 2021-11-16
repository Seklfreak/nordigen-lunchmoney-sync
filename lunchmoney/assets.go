package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Asset represents a single asset.
type Asset struct {
	ID        int        `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// parameters to update.
	TypeName        string     `json:"type_name,omitempty"`
	SubtypeName     string     `json:"subtype_name,omitempty"`
	Name            string     `json:"name,omitempty"`
	Balance         string     `json:"balance,omitempty"`
	BalanceAsOf     *time.Time `json:"balance_as_of,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	InstitutionName string     `json:"institution_name,omitempty"`
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
		Error []string `json:"error"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if len(result.Error) > 0 {
		return fmt.Errorf("received %d errors: %q", len(result.Error), strings.Join(result.Error, "; "))
	}

	return nil
}
