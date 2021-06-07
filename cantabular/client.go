package cantabular

import (
	"net/http"
	"errors"
	"time"
	"fmt"
	"encoding/json"
)

// Client is the client for interacting with the Cantabular API
type Client struct{
	ua *http.Client
	host string
}

// NewClient returns a new Client
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

// errorResponse handles dealing with an error response from Cantabular
func (c *Client) errorResponse(res *http.Response) error{
	var resp ErrorResponse

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil{
		return &Error{
			err: fmt.Errorf("failed to decode error response body: %s", err),
			statusCode: res.StatusCode,
		}
	}

	return &Error{
		err: errors.New(resp.Message),
		statusCode: res.StatusCode,
	}
}