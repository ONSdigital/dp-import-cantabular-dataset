package recipe

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"fmt"
)

// Client provides a client for calling the Recipe API.
type Client struct {
	ua        httpClient
	host      string
}

// NewClient returns a new Client
func NewClient(ua httpClient, cfg Config) *Client{
	return &Client{
		ua:   ua,
		host: cfg.Host,
	}
}

// errorResponse handles dealing with an error response from the Recipe API
func (c *Client) errorResponse(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return &Error{
			err: fmt.Errorf("failed to read error response body: %s", err),
			statusCode: res.StatusCode,
		}
	}

	if len(b) == 0{
		b = []byte("[response body empty]")
	}

	return &Error{
		err: fmt.Errorf("error response from recipe-api: %v", string(b)),
		statusCode: res.StatusCode,
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
		return nil, fmt.Errorf("failed to make request: %s", err)
	}

	return resp, nil
}
