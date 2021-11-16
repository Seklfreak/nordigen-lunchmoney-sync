package lunchmoney

import (
	"bytes"
	"context"
	"net/http"

	"github.com/pkg/errors"
)

const baseURL = "https://dev.lunchmoney.app"

// Client allows interacting with the Lunchmoney API.
type Client struct {
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new client.
func NewClient(accessToken string, httpClient *http.Client) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient:  httpClient,
	}
}

func (c *Client) createRequest(ctx context.Context, method string, endpoint string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	return req, nil
}
