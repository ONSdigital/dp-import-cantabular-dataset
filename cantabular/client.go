package cantabular

import (
	"net/http"
)

// Client is the client for interacting with the Cantabular API
type Client struct{
	ua *http.Client
	host string
}

func NewClient(cfg Config) *Client{
	return &Client{
		ua: &http.Client{
			Timeout: cfg.Timeout,
		},
		host: cfg.Host,
	}
}