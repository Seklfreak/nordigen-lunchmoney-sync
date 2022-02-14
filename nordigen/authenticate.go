package nordigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// authenticate fetches the access key using the Secret ID and Secret Key.
func (c *Client) authenticate(ctx context.Context) (string, error) {
	reqBody := struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}{
		SecretID:  c.config.SecretID,
		SecretKey: c.config.SecretKey,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/token/new/", baseURL),
		bytes.NewReader(reqBodyBytes),
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to create http request")
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to make http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err := extractError(resp); err != nil {
			return "", errors.Wrap(err, "failed to authenticate")
		}

		return "", errors.Errorf("received unexpected status code when authenticating: %s", resp.Status)
	}

	var creds struct {
		Access string `json:"access"`
	}

	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode response body")
	}

	return creds.Access, nil
}
