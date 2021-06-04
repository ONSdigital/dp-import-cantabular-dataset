package cantabular

import (
	"context"
	"fmt"
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

	url := fmt.Sprintf("%s/codebook/%s?cats=%v%s", c.host, req.DatasetName, req.Categories, vars)

	fmt.Println("making requests to ", url)

	return nil, nil
}