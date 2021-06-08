package cantabular

import (
	"net/http"
	"errors"
	"context"
	"net/url"
	"fmt"
	"encoding/json"
)

// Client is the client for interacting with the Cantabular API
type Client struct{
	ua   httpClient
	host string
}

// NewClient returns a new Client
func NewClient(ua httpClient, cfg Config) *Client{
	return &Client{
		ua:   ua,
		host: cfg.Host,
	}
}

// get makes a get request to the given url and returns the response
func (c *Client) httpGet(ctx context.Context, path string) (*http.Response, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, &Error{
			err: fmt.Errorf("failed to parse url: %s", err),
			statusCode: http.StatusBadRequest,
		}
	}

	path = URL.String()

	resp, err := c.ua.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// errorResponse handles dealing with an error response from Cantabular
func (c *Client) errorResponse(res *http.Response) error {
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
