package cantabular

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
