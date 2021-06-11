package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-import-cantabular-dataset/models"
)

type cantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type datasetAPIClient interface{
	PutDataset(context.Context, string, string, string, string, dataset.DatasetDetails) error
}

type recipeAPIClient interface{
	Get(context.Context, string) (*models.Recipe, error)
}