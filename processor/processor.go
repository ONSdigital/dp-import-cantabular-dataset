package processor

import (
	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
)

type Processor struct{
	importAPI  ImportAPIClient
	datasetAPI DatasetAPIClient
	cfg config.Config
}

func New(cfg config.Config, i ImportAPIClient, d DatasetAPIClient) *Processor{
	return &Processor{
		cfg:        cfg,
		importAPI:  i,
		datasetAPI: d,
	}
}