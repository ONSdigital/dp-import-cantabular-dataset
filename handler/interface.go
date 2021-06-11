package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
)

type cantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type datasetAPIClient interface{
	PutDataset(context.Context, string, string, string, string, dataset.DatasetDetails) error
}

type recipeAPIClient interface{
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
}