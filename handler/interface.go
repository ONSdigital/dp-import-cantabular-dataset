package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
)

type cantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type datasetAPIClient interface{
	PutInstance(context.Context, string, string, string, string, dataset.UpdateInstance) error
}

type recipeAPIClient interface{
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
}