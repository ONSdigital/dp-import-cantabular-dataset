package handler

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
)

//go:generate moq -out mock/cantabular_client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset_api_client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/recipe_api_client.go -pkg mock . RecipeAPIClient
//go:generate moq -out mock/import_api_client.go -pkg mock . ImportAPIClient

type CantabularClient interface {
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
}

type DatasetAPIClient interface {
	PutInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string, i dataset.UpdateInstance, ifMatch string) (eTag string, err error)
	PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state dataset.State, ifMatch string) (eTag string, err error)
}

type RecipeAPIClient interface {
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
}

type ImportAPIClient interface {
	UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState importapi.State) error
}

type dataLogger interface {
	LogData() map[string]interface{}
}

type instanceCompleteder interface {
	InstanceCompleted() bool
}
