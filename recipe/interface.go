package recipe

import (
	"net/http"
	"context"
)

// httpClient is an interface for a user agent to make http requests
type httpClient interface{
	Get(ctx context.Context, url string) (*http.Response, error)
}

type coder interface{
	Code() int
}