package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-import-cantabular-dataset/cantabular"
)

type CantabularClient interface{
	GetCodebook(cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type DatasetAPIClient interface{
	PutDataset(context.Context, string, string, string, string, dataset.DatasetDetails) error
}