package nordigen

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

const baseURL = "https://ob.nordigen.com/api/v2"

// Config is the configuration for the Nordigen client.
type Config struct {
	SecretID  string `envconfig:"SECRET_ID"`
	SecretKey string `envconfig:"SECRET_KEY"`
}

// Client represents a Nordigen API client.
type Client struct {
	config     *Config
	httpClient *http.Client

	accessKey string
}

// NewClient creates a new Nordigen API client.
func NewClient(config *Config, httpClient *http.Client) (*Client, error) {
	if config == nil || config.SecretID == "" || config.SecretKey == "" {
		return nil, errors.New("invalid config")
	}

	// create client
	client := &Client{
		config:     config,
		httpClient: httpClient,
	}

	// authenticate client
	accessKey, err := client.authenticate(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to authenticate")
	}
	client.accessKey = accessKey

	return client, nil
}
