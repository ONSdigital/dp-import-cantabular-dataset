package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/recipe"

)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/health-check.go -pkg mock . HealthChecker
//go:generate moq -out mock/canabular-client.go -pkg mock . CantabularClient
//go:generate moq -out mock/dataset-api-client.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/recipe-api-client.go -pkg mock . RecipeAPIClient

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
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

type CantabularClient interface{
	GetCodebook(context.Context, cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)
	Checker(context.Context, *healthcheck.CheckState) error
}

type DatasetAPIClient interface{
	// Not sure which calls we will be making yet
	Checker(context.Context, *healthcheck.CheckState) error
}

type RecipeAPIClient interface{
	GetRecipe(context.Context, string, string, string) (*recipe.Recipe, error)
	Checker(context.Context, *healthcheck.CheckState) error
}
