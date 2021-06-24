package processor

type Processor struct{
	numWorkers int
	importAPI  ImportAPIClient
	datasetAPI DatasetAPIClient
}

func New(n int, i ImportAPIClient, d DatasetAPIClient) *Processor{
	return &Processor{
		numWorkers: n,
		importAPI:  i,
		datasetAPI: d,
	}
}