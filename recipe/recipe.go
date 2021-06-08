package recipe

import(
	"context"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-import-cantabular-dataset/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// Get Recipe from an ID
func (c *Client) Get(ctx context.Context, id string) (*models.Recipe, error) {
	url := fmt.Sprintf("%s/recipes/%s", c.host, id)

	log.Info(ctx, "Getting recipe from recipe-api", log.Data{"recipe_id": id, "url": url})

	res, err := c.httpGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to call recipe api: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, c.errorResponse(res)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, &Error{
			err: fmt.Errorf("failed to read response from recipe-api: %s", err),
			statusCode: res.StatusCode,
		}
	}

	var r models.Recipe
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, &Error{
			err: fmt.Errorf("failed to unmarshal response from recipe-api: %s", err),
			statusCode: res.StatusCode,
			logData: map[string]interface{}{
				"response_body": string(b),
			},
		}
	}

	return &r, nil
}