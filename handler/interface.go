package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
)

type cantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type datasetAPIClient interface{
	// Not sure which calls we will be making yet
}

type recipeAPIClient interface{
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
}