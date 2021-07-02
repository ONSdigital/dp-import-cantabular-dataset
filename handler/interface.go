package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
)

//go:generate moq -out mock/cantabular_client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset_api_client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/recipe_api_client.go -pkg mock . RecipeAPIClient

type CantabularClient interface {
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type DatasetAPIClient interface {
	PutInstance(context.Context, string, string, string, string, dataset.UpdateInstance) error
	PutInstanceState(context.Context, string, string, dataset.State) error
}

type RecipeAPIClient interface {
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
}

type dataLogger interface {
	LogData() map[string]interface{}
}
