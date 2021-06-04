package cantabular

import (
	"net/http"
	"time"
)

// Client is the client for interacting with the Cantabular API
type Client struct{
	ua *http.Client
	host string
}

func NewClient(cfg Config) *Client{
	var to time.Duration
	var err error

	if to, err = time.ParseDuration(cfg.Timeout); err != nil{
		to = time.Second * 30
	}

	return &Client{
		ua: &http.Client{
			Timeout: to,
		},
		host: cfg.Host,
	}
}