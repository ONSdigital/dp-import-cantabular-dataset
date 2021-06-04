package cantabular

import (
	"context"
	"fmt"
	"encoding/json"
)

// Variable represents a 'codebook' object returned from Cantabular
type Codebook []Variable

type GetCodebookRequest struct{
	DatasetName string 
	Variables   []string
	Categories  bool
}

type GetCodebookResponse struct{
	Codebook Codebook `json:"codebook"`
	Dataset  Dataset  `json:"dataset"`
}

func (c *Client) GetCodebook(ctx context.Context, req GetCodebookRequest) (*GetCodebookResponse, error){
	var vars string
	for _, v := range req.Variables{
		vars += "&v=" + v
	}

	url := fmt.Sprintf("%s/v8/codebook/%s?cats=%v%s", c.host, req.DatasetName, req.Categories, vars)

	fmt.Println("making requests to ", url)

	res, err := c.ua.Get(url)
	if err != nil{
		return nil, fmt.Errorf("failed to get response from Cantabular API: %s", err)
	}

	defer res.Body.Close()

	var resp GetCodebookResponse

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil{
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &resp, nil
}