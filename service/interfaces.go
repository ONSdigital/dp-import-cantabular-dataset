package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/health_check.go -pkg mock . HealthChecker
//go:generate moq -out mock/canabular_client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset_api_client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/recipe_api_client.go -pkg mock . RecipeAPIClient

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddAndGetCheck(name string, checker healthcheck.Checker) (check *healthcheck.Check, err error)
	Subscribe(s healthcheck.Subscriber, checks ...*healthcheck.Check)
}

type CantabularClient interface {
	GetDimensionsByName(ctx context.Context, req cantabular.StaticDatasetQueryRequest) (*cantabular.GetDimensionsResponse, error)
	Checker(context.Context, *healthcheck.CheckState) error
	CheckerAPIExt(ctx context.Context, state *healthcheck.CheckState) error
}

type DatasetAPIClient interface {
	PutInstance(context.Context, string, string, string, string, dataset.UpdateInstance, string) (string, error)
	PutInstanceState(context.Context, string, string, dataset.State, string) (string, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

type RecipeAPIClient interface {
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

type ImportAPIClient interface {
	UpdateImportJobState(context.Context, string, string, importapi.State) error
}
