package cantabular

import (
	"context"
	"fmt"
	"net/http"
	"encoding/json"
)

// Variable represents a 'codebook' object returned from Cantabular
type Codebook []Variable

// GetCodebook gets a Codebook from cantabular.
// TODO: Should this return the entire CodebookResponse or just the Codebook?
func (c *Client) GetCodebook(ctx context.Context, req GetCodebookRequest) (*GetCodebookResponse, error){
	var vars string
	for _, v := range req.Variables{
		vars += "&v=" + v
	}

	url := fmt.Sprintf("%s/v8/codebook/%s?cats=%v%s", c.host, req.DatasetName, req.Categories, vars)

	fmt.Println("making requests to ", url)

	res, err := c.ua.Get(url)
	if err != nil{
		return nil, &Error{
			err: fmt.Errorf("failed to get response from Cantabular API: %s", err),
			statusCode: http.StatusInternalServerError,
		}
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK{
		return nil, c.errorResponse(res)
	}

	var resp GetCodebookResponse

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil{
		return nil, &Error{
			err: fmt.Errorf("failed to decode response body: %s", err),
			statusCode: http.StatusInternalServerError,
		}
	}

	return &resp, nil
}