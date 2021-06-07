package cantabular

import (
	"encoding/json"
	"fmt"
)

// ErrorResponse models the error response from cantabular
type ErrorResponse struct{
	Message string `json:"message"`
}

type GetCodebookRequest struct{
	DatasetName string 
	Variables   []string
	Categories  bool
}

type GetCodebookResponse struct{
	Codebook Codebook `json:"codebook"`
	Dataset  Dataset  `json:"dataset"`
}

// temp function for dev
func (r GetCodebookResponse) String() string{
	b, err := json.MarshalIndent(&r, "", "    ")
	if err != nil{
		return fmt.Sprintf("%+v", r)
	}

	return string(b)
}